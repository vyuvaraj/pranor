package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

const usage = `ServStore CLI — unified command-line tool for ServStore object storage and cluster administration.

Usage:
  servstore [global flags] <command> [args]

Global flags:
  --endpoint       ServStore S3 API URL        (default: http://localhost:9000)
  --admin-endpoint ServStore Admin API URL     (default: http://localhost:9001)
  --access-key     Access key ID              (default: minioadmin)
  --secret-key     Secret access key          (default: minioadmin)

S3 & Data Commands:
  mb   <bucket>                      Create a bucket
  rb   <bucket>                      Remove a bucket
  ls   [bucket [prefix]]             List buckets or objects in a bucket
  put  <bucket> <key> <file>         Upload a file as an object
  get  <bucket> <key> <dest>         Download an object to a file (use - for stdout)
  rm   <bucket> <key>                Delete an object
  lock <bucket> <key> <duration>     Apply a WORM lock (e.g. 72h, 30d, 1y)
  lc-set     <bucket> <days> [prefix] Set lifecycle expiry rule (delete objects older than N days)
  lc-get     <bucket>                 Show the lifecycle configuration for a bucket
  lc-del     <bucket>                 Remove the lifecycle configuration from a bucket
  policy-set <username> <file.json>   Attach an IAM policy to a user
  policy-get <username>               View the attached IAM policy for a user
  policy-del <username>               Delete the attached IAM policy for a user
  cluster-status                      View the status of all nodes in the cluster
  placement  <bucket> <key>          Find the node owning a specific key

Admin & Daemon Commands:
  status                              Check status of running servstored daemon
  admin-buckets                       List buckets via admin API
  admin-create-bucket <bucket>        Create a bucket via admin API
  version                             Print version information
  help                                Print this message
`

// ---------- global flags ----------

var (
	endpoint      string
	adminEndpoint string
	accessKey     string
	secretKey     string
)

func main() {
	flag.StringVar(&endpoint, "endpoint", "http://localhost:9000", "ServStore S3 API URL")
	flag.StringVar(&adminEndpoint, "admin-endpoint", "http://localhost:9001", "ServStore Admin API URL")
	flag.StringVar(&accessKey, "access-key", "minioadmin", "Access key ID")
	flag.StringVar(&secretKey, "secret-key", "minioadmin", "Secret access key")
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Print(usage)
		os.Exit(0)
	}

	cmd := args[0]
	rest := args[1:]

	var err error
	switch cmd {
	case "mb":
		err = cmdMB(rest)
	case "rb":
		err = cmdRB(rest)
	case "ls":
		err = cmdLS(rest)
	case "put":
		err = cmdPut(rest)
	case "get":
		err = cmdGet(rest)
	case "rm":
		err = cmdRM(rest)
	case "lock":
		err = cmdLock(rest)
	case "lc-set":
		err = cmdLCSet(rest)
	case "lc-get":
		err = cmdLCGet(rest)
	case "lc-del":
		err = cmdLCDel(rest)
	case "policy-set":
		err = cmdPolicySet(rest)
	case "policy-get":
		err = cmdPolicyGet(rest)
	case "policy-del":
		err = cmdPolicyDel(rest)
	case "cluster-status":
		err = cmdClusterStatus(rest)
	case "placement":
		err = cmdPlacement(rest)
	case "status":
		err = cmdAdminStatus(rest)
	case "admin-buckets":
		err = cmdAdminBuckets(rest)
	case "admin-create-bucket":
		err = cmdAdminCreateBucket(rest)
	case "version":
		fmt.Println("servstore CLI v2.0.0 (Unified Data & Admin Tool)")
	case "help", "--help", "-h":
		fmt.Print(usage)
	default:
		fatalf("unknown command %q — run 'servstore help' for usage\n", cmd)
	}

	if err != nil {
		fatalf("%v\n", err)
	}
}

// ---------- commands ----------

func cmdMB(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mb <bucket>")
	}
	bucket := args[0]
	req, _ := http.NewRequest(http.MethodPut, url("/"+bucket), nil)
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK, http.StatusNoContent); err != nil {
		return err
	}
	fmt.Printf("Bucket '%s' created.\n", bucket)
	return nil
}

func cmdRB(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: rb <bucket>")
	}
	bucket := args[0]
	req, _ := http.NewRequest(http.MethodDelete, url("/"+bucket), nil)
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK, http.StatusNoContent); err != nil {
		return err
	}
	fmt.Printf("Bucket '%s' removed.\n", bucket)
	return nil
}

func cmdLS(args []string) error {
	if len(args) == 0 {
		return listAllBuckets()
	}
	bucket := args[0]
	prefix := ""
	if len(args) > 1 {
		prefix = args[1]
	}
	return listObjects(bucket, prefix)
}

func listAllBuckets() error {
	req, _ := http.NewRequest(http.MethodGet, url("/"), nil)
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return err
	}

	var res struct {
		XMLName xml.Name `xml:"ListAllMyBucketsResult"`
		Buckets struct {
			Bucket []struct {
				Name         string `xml:"Name"`
				CreationDate string `xml:"CreationDate"`
			} `xml:"Bucket"`
		} `xml:"Buckets"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := xml.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("decode xml: %w", err)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "CREATION DATE\tBUCKET")
	for _, b := range res.Buckets.Bucket {
		fmt.Fprintf(tw, "%s\t%s\n", formatDate(b.CreationDate), b.Name)
	}
	return tw.Flush()
}

func listObjects(bucket, prefix string) error {
	u := url("/" + bucket)
	if prefix != "" {
		u += "?prefix=" + prefix
	}
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return err
	}

	var res struct {
		XMLName  xml.Name `xml:"ListBucketResult"`
		Name     string   `xml:"Name"`
		Contents []struct {
			Key          string `xml:"Key"`
			LastModified string `xml:"LastModified"`
			Size         int64  `xml:"Size"`
			ETag         string `xml:"ETag"`
		} `xml:"Contents"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := xml.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("decode xml: %w", err)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "LAST MODIFIED\tSIZE\tETAG\tKEY")
	for _, c := range res.Contents {
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\n", formatDate(c.LastModified), c.Size, c.ETag, c.Key)
	}
	return tw.Flush()
}

func cmdPut(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: put <bucket> <key> <file>")
	}
	bucket, key, file := args[0], args[1], args[2]

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open file %s: %w", file, err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat file %s: %w", file, err)
	}

	req, _ := http.NewRequest(http.MethodPut, url("/"+bucket+"/"+key), f)
	req.ContentLength = fi.Size()
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK, http.StatusCreated); err != nil {
		return err
	}
	fmt.Printf("Uploaded '%s' → s3://%s/%s (%d bytes).\n", file, bucket, key, fi.Size())
	return nil
}

func cmdGet(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: get <bucket> <key> <dest>")
	}
	bucket, key, dest := args[0], args[1], args[2]

	req, _ := http.NewRequest(http.MethodGet, url("/"+bucket+"/"+key), nil)
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return err
	}

	var out io.Writer
	if dest == "-" {
		out = os.Stdout
	} else {
		f, err := os.Create(dest)
		if err != nil {
			return fmt.Errorf("create dest %s: %w", dest, err)
		}
		defer f.Close()
		out = f
	}

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	if dest != "-" {
		fmt.Printf("Downloaded s3://%s/%s → '%s' (%d bytes).\n", bucket, key, dest, n)
	}
	return nil
}

func cmdRM(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: rm <bucket> <key>")
	}
	bucket, key := args[0], args[1]
	req, _ := http.NewRequest(http.MethodDelete, url("/"+bucket+"/"+key), nil)
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK, http.StatusNoContent); err != nil {
		return err
	}
	fmt.Printf("Removed s3://%s/%s\n", bucket, key)
	return nil
}

func cmdLock(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: lock <bucket> <key> <duration>")
	}
	bucket, key, durationStr := args[0], args[1], args[2]

	dur, err := parseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", durationStr, err)
	}

	retainUntil := time.Now().Add(dur).Format(time.RFC3339)

	req, _ := http.NewRequest(http.MethodPut, url("/"+bucket+"/"+key+"?retention"), nil)
	req.Header.Set("X-Amz-Object-Lock-Retain-Until-Date", retainUntil)
	req.Header.Set("X-Amz-Object-Lock-Mode", "COMPLIANCE")

	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK, http.StatusNoContent); err != nil {
		return err
	}
	fmt.Printf("Locked s3://%s/%s until %s (mode: COMPLIANCE)\n", bucket, key, retainUntil)
	return nil
}

func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	if strings.HasSuffix(s, "y") {
		years, err := strconv.Atoi(strings.TrimSuffix(s, "y"))
		if err != nil {
			return 0, err
		}
		return time.Duration(years) * 365 * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func cmdLCSet(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: lc-set <bucket> <days> [prefix]")
	}
	bucket := args[0]
	daysStr := args[1]
	prefix := ""
	if len(args) > 2 {
		prefix = args[2]
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		return fmt.Errorf("invalid days %q: must be a positive integer", daysStr)
	}

	xmlPayload := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<LifecycleConfiguration>
  <Rule>
    <ID>rule-1</ID>
    <Prefix>%s</Prefix>
    <Status>Enabled</Status>
    <Expiration>
      <Days>%d</Days>
    </Expiration>
  </Rule>
</LifecycleConfiguration>`, prefix, days)

	req, _ := http.NewRequest(http.MethodPut, url("/"+bucket+"?lifecycle"), strings.NewReader(xmlPayload))
	req.Header.Set("Content-Type", "application/xml")
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK, http.StatusNoContent); err != nil {
		return err
	}
	fmt.Printf("Lifecycle rule set for bucket '%s': expire objects older than %d days (prefix: %q)\n", bucket, days, prefix)
	return nil
}

func cmdLCGet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: lc-get <bucket>")
	}
	bucket := args[0]
	req, _ := http.NewRequest(http.MethodGet, url("/"+bucket+"?lifecycle"), nil)
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return err
	}
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	return nil
}

func cmdLCDel(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: lc-del <bucket>")
	}
	bucket := args[0]
	req, _ := http.NewRequest(http.MethodDelete, url("/"+bucket+"?lifecycle"), nil)
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK, http.StatusNoContent); err != nil {
		return err
	}
	fmt.Printf("Lifecycle configuration removed from bucket '%s'\n", bucket)
	return nil
}

func cmdPolicySet(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: policy-set <username> <file.json>")
	}
	username, file := args[0], args[1]
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read policy file %s: %w", file, err)
	}

	req, _ := http.NewRequest(http.MethodPut, url("/admin/policy/"+username), bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK, http.StatusNoContent); err != nil {
		return err
	}
	fmt.Printf("Policy set for user %q from %s\n", username, file)
	return nil
}

func cmdPolicyGet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: policy-get <username>")
	}
	username := args[0]
	req, _ := http.NewRequest(http.MethodGet, url("/admin/policy/"+username), nil)
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return err
	}
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	return nil
}

func cmdPolicyDel(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: policy-del <username>")
	}
	username := args[0]
	req, _ := http.NewRequest(http.MethodDelete, url("/admin/policy/"+username), nil)
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK, http.StatusNoContent); err != nil {
		return err
	}
	fmt.Printf("Policy removed for user %q\n", username)
	return nil
}

func cmdClusterStatus(args []string) error {
	req, _ := http.NewRequest(http.MethodGet, url("/admin/cluster/status"), nil)
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return err
	}

	var nodes []struct {
		ID      string `json:"id"`
		Address string `json:"address"`
		Role    string `json:"role"`
		Status  string `json:"status"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &nodes); err != nil {
		fmt.Println(string(body))
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NODE ID\tADDRESS\tROLE\tSTATUS")
	for _, n := range nodes {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", n.ID, n.Address, n.Role, n.Status)
	}
	return tw.Flush()
}

func cmdPlacement(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: placement <bucket> <key>")
	}
	bucket, key := args[0], args[1]
	req, _ := http.NewRequest(http.MethodGet, url("/admin/placement/"+bucket+"/"+key), nil)
	resp, err := do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := expectStatus(resp, http.StatusOK); err != nil {
		return err
	}

	var res struct {
		Bucket  string `json:"bucket"`
		Key     string `json:"key"`
		NodeID  string `json:"node_id"`
		Address string `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("decode placement: %w", err)
	}

	fmt.Printf("Bucket:     %s\n", res.Bucket)
	fmt.Printf("Key:        %s\n", res.Key)
	fmt.Printf("Owner Node: %s (%s)\n", res.NodeID, res.Address)
	return nil
}

func cmdAdminStatus(args []string) error {
	adminURL := strings.TrimRight(adminEndpoint, "/") + "/api/v1/health"
	resp, err := client.Get(adminURL)
	if err != nil {
		return fmt.Errorf("error connecting to servstored daemon at %s: %w", adminURL, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var statusMap map[string]interface{}
	if err := json.Unmarshal(body, &statusMap); err != nil {
		return fmt.Errorf("invalid status JSON: %w", err)
	}

	fmt.Println("ServStore Daemon Status:")
	for k, v := range statusMap {
		fmt.Printf("  %-15s: %v\n", k, v)
	}
	return nil
}

func cmdAdminBuckets(args []string) error {
	adminURL := strings.TrimRight(adminEndpoint, "/") + "/api/v1/buckets"
	resp, err := client.Get(adminURL)
	if err != nil {
		return fmt.Errorf("error listing buckets via admin API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Active Storage Buckets (Admin API):")
	fmt.Println(string(body))
	return nil
}

func cmdAdminCreateBucket(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: admin-create-bucket <bucket_name>")
	}
	bucketName := args[0]
	payload, _ := json.Marshal(map[string]string{
		"name": bucketName,
	})

	adminURL := strings.TrimRight(adminEndpoint, "/") + "/api/v1/buckets"
	resp, err := client.Post(adminURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating bucket via admin API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		fmt.Printf("Successfully created bucket via admin API: %s\n", bucketName)
	} else {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create bucket: %s", string(body))
	}
	return nil
}

// ---------- helpers ----------

var client = &http.Client{Timeout: 30 * time.Second}

func url(path string) string {
	return strings.TrimRight(endpoint, "/") + path
}

func do(req *http.Request) (*http.Response, error) {
	if accessKey != "" && secretKey != "" {
		req.SetBasicAuth(accessKey, secretKey)
	}
	return client.Do(req)
}

func expectStatus(resp *http.Response, codes ...int) error {
	for _, c := range codes {
		if resp.StatusCode == c {
			return nil
		}
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

func formatDate(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("2006-01-02 15:04:05")
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format, a...)
	os.Exit(1)
}
