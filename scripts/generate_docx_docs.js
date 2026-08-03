const fs = require('fs');
const path = require('path');
const docx = require('docx');
const { Document, Packer, Paragraph, TextRun, HeadingLevel, Table, TableRow, TableCell, WidthType, ShadingType } = docx;

// Colors
const COLOR_PRIMARY = "1E3A8A"; // Dark Blue
const COLOR_SECONDARY = "0D9488"; // Teal
const COLOR_DARK = "1F2937";

const modulesData = [
  {
    name: "Pranor Gate",
    subtitle: "AI-Native API Gateway & Ingress Controller",
    overview: "Pranor Gate is an ultra-high performance, eBPF-accelerated, AI-native API Gateway and ingress controller built for modern microservices and zero-trust cloud architectures. It delivers sub-millisecond HTTP/gRPC proxying, WebAssembly policy execution, semantic prompt security, and active-active multi-region failover.",
    components: [
      { title: "eBPF XDP Kernel Bypass DDoS Protection (EE.85.1)", desc: "Intercepts inbound IP packets directly in the network card driver layer (XDP_DROP/XDP_PASS) before memory allocation in the Linux socket stack. Handles 100Gbps line-rate packet volumes under active volumetric DDoS attacks with < 1% CPU overhead." },
      { title: "AI Prompt Firewall & MCP Tool Gateway (EE.85.146, EE.87.9)", desc: "Real-time evaluation of prompt embeddings against known adversarial injection attack patterns (jailbreaks, prompt overrides). Inline regex and NER inspection auto-redacting SSNs, credit cards, and API keys before forwarding prompts to LLMs. Native support for Model Context Protocol (MCP) JSON-RPC tool routing." },
      { title: "Global CRDT Rate-Limiting Grid (EE.86.7)", desc: "Uses Conflict-Free Replicated Data Types (PN-Counters) over an encrypted gossip protocol for sub-millisecond global quota synchronization across multi-region edge clusters." },
      { title: "GPU/TPU Accelerator Offloader & WASM Sandboxing (EE.87.6, EE.87.21)", desc: "Routes high-throughput LLM token streams via direct PCIe kernel bypass to hardware accelerators (NVIDIA H100 / Google TPU). Enforces memory-bound WebAssembly execution runtime for custom extensions." }
    ],
    matrix: [
      ["Ingress Proxy", "Sub-millisecond HTTP/1.1, HTTP/2, gRPC routing", "PCIe hardware TLS offloading & 100Gbps eBPF XDP bypass"],
      ["WASM Hot-Swap", "Zero-downtime WebAssembly middleware", "Strict memory-bound WASM runtime security sandboxing"],
      ["AI Agent (MCP) Gateway", "Native MCP JSON-RPC routing & cost headers", "Semantic prompt firewall, PII redaction & injection guard"],
      ["DDoS Defense", "Token bucket sliding-window rate limiting", "Kernel eBPF XDP packet filtering at 100Gbps line rate"],
      ["Multi-Region Failover", "Dynamic DNS & canary promotion", "1-click multi-region DR failover & Anycast steering"]
    ],
    catalog: [
      "EE.85.1: Kernel eBPF XDP DDoS Bypass (100Gbps line-rate packet dropping)",
      "EE.85.2: Geo-IP Latency Anycast Steering (Real-time edge routing to lowest-latency datacenter)",
      "EE.85.3: GraphQL Schema Stitching & Federation (Unified backend GraphQL schema gateway)",
      "EE.86.7: Global CRDT Rate-Limiting Grid (Sub-millisecond global budget synchronization)",
      "EE.86.8: 1-Click Multi-Region Active-Passive DR Failover (Automated DNS failover)",
      "EE.86.16: Continuous Threat Intelligence Feed Sync (Real-time IP reputation ingestion)",
      "EE.87.6: Hardware GPU/TPU Accelerator Offloader (Direct PCIe token stream routing)",
      "EE.87.9: AI Prompt Injection Guard & Poison Pill Sanitizer (Adversarial prompt filtering)"
    ]
  },
  {
    name: "Pranor Vault",
    subtitle: "High-Performance Object Storage & HNSW Vector Engine",
    overview: "Pranor Vault is an enterprise-grade object storage server and vector database engine. It seamlessly unifies AWS S3 API compatibility, high-dimensional vector similarity search, zero-knowledge homomorphic search, sovereign geofenced isolation, and SEC 17a-4 compliant WORM retention locking.",
    components: [
      { title: "S3 Compatibility & Native HNSW Vector Engine", desc: "Full support for standard S3 SDKs, MinIO CLI, bucket lifecycle policies, and multipart uploads. In-memory Hierarchical Navigable Small World (HNSW) graphs delivering sub-5ms similarity search over 100M+ vector embeddings." },
      { title: "AI Sovereign Vector Isolation & Homomorphic Search (EE.87.1, EE.87.2)", desc: "Enforces strict physical datacenter data boundaries for vector indices to guarantee cross-border sovereignty compliance (EU GDPR / US HIPAA). Computes similarity distances directly over encrypted vector payloads." },
      { title: "Copy-on-Write (CoW) Instant Bucket Branching (EE.85.6)", desc: "Instant git-like bucket branching ('pranor-vault branch create sandbox') enabling sandbox development over petabyte storage buckets without data duplication." },
      { title: "GDPR Right-to-be-Forgotten Purge Worker & WORM Vault (EE.87.3, EE.85.4)", desc: "Automated cryptographic record destruction sweeping WAL logs and vector indices to fulfill GDPR / CCPA erasure requests with verifiable proofs. SEC 17a-4 compliant WORM retention locks." }
    ],
    matrix: [
      ["S3 API & Vector Search", "Full S3 compatibility & HNSW vector search", "Sovereign geofenced vector isolation & homomorphic search"],
      ["Bucket Management", "Standard bucket operations & policies", "Copy-on-Write (CoW) instant petabyte bucket branching"],
      ["Privacy & Compliance", "Client-side envelope encryption", "Automated GDPR zeroization purge worker & SEC WORM lock"],
      ["Data Protection", "In-flight TLS & AES-256 static encryption", "Dynamic PII masking & Zero-Knowledge MPC threshold keys"]
    ],
    catalog: [
      "EE.85.4: Immutable Object Access Audit Trail (Append-only identity log for object reads/writes)",
      "EE.85.5: Active-Active Multi-Region Replication (Cross-cloud replication with LWW conflict resolution)",
      "EE.85.6: Copy-on-Write (CoW) Bucket Branching (Instant sandbox environment cloning)",
      "EE.87.1: AI Sovereign Vector Isolation & Data Boundary Enforcer (Geofenced index isolation)",
      "EE.87.2: Zero-Knowledge Homomorphic Search Engine (Encrypted similarity distance computation)",
      "EE.87.3: Automated GDPR / CCPA Right-to-be-Forgotten Purge Worker (Cryptographic zeroization)"
    ]
  },
  {
    name: "Pranor Pulse",
    subtitle: "Ultra-Low Latency Multi-Protocol Event Broker & Streaming WAL",
    overview: "Pranor Pulse is an ultra-low latency event broker and streaming messaging queue designed for mission-critical enterprise event-driven architectures. It unifies wire-level support for Kafka, STOMP, and MQTT with SIMD event filtering, 2PC atomic transactions, and active-active topic mirroring.",
    components: [
      { title: "Multi-Protocol Engine & Zero-Copy Hardware WAL (EE.86.10)", desc: "Native wire protocol translators for Kafka producers, STOMP WebSockets, and MQTT 3.1/5.0 IoT devices. Hardware-accelerated zero-copy Write-Ahead Logging (WAL) with AES-NI encryption." },
      { title: "Active-Active Cross-Cloud MirrorMaker v2 (EE.86.9)", desc: "Continuous active-active topic replication across AWS MSK, GCP PubSub, and Azure Event Hubs with poison-pill message filtration and AI consumer group rebalancing." },
      { title: "SIMD AVX-512 Event Filtering & Blind Broker Encryption (EE.86.18, EE.86.17)", desc: "Filters event payload headers and bodies using vectorized SIMD CPU instructions at wire speed (1M+ events/sec). Zero-trust end-to-end payload encryption." },
      { title: "Exactly-Once 2PC Transaction Coordinator (EE.87.10)", desc: "Distributed atomic transaction coordinator enforcing strict exactly-once delivery guarantees across multi-broker clusters." }
    ],
    matrix: [
      ["Multi-Protocol Engine", "Kafka, STOMP, and MQTT wire protocol support", "Multi-tenant partition memory sharding"],
      ["Delivery Guarantees", "Sliding-window deduplication & DLQ replay", "Two-Phase Commit (2PC) exactly-once transaction coordinator"],
      ["Replication", "Single-cluster leader-follower replication", "Active-active cross-cloud MirrorMaker v2 topic sync"],
      ["Event Filtering", "Declarative WASM transform pipeline", "SIMD AVX-512 hardware-accelerated payload filter engine"]
    ],
    catalog: [
      "EE.85.8: Multi-Region MirrorMaker Sync (Active-active cross-cloud event topic mirroring)",
      "EE.85.9: Hardware Payload Encryption at Rest (KMS/HSM envelope encryption before WAL write)",
      "EE.86.9: Active-Active Cross-Cloud MirrorMaker v2 (Poison-pill filtering cross-cloud sync)",
      "EE.86.10: Hardware-Accelerated Zero-Copy WAL Encryption (Hardware AES-NI disk encryption)",
      "EE.86.17: Blind Broker End-to-End Payload Encryption (Zero-trust payload protection)",
      "EE.87.10: Distributed Multi-Broker Transaction Coordinator (2PC atomic manager)"
    ]
  },
  {
    name: "Pranor Flow",
    subtitle: "Durable Saga Workflow Orchestrator & Distributed Execution Engine",
    overview: "Pranor Flow is a resilient, code-first workflow orchestrator for long-running business processes, microservice sagas, and AI agent execution pipelines. It guarantees deterministic state recovery, compensation handling, and temporal execution state persistence.",
    components: [
      { title: "Distributed Saga State Engine & Compensation Ordering", desc: "Guarantees backward-rollback compensation step execution during microservice failures. Persists exact step execution state in durable WAL storage." },
      { title: "AI Agent Orchestration & Step Replay", desc: "Native support for complex multi-step LLM chain workflows with automatic state snapshotting and cost-aware retry backoffs." },
      { title: "BFT Distributed Consensus State Manager (EE.87.7)", desc: "Byzantine Fault Tolerant (BFT) Raft consensus ensuring workflow state integrity even under compromised cluster nodes." },
      { title: "Active-Active Saga Migration & State Mirroring (EE.86.11)", desc: "Zero-downtime saga state migration allowing active workflows to migrate smoothly across cloud regions." }
    ],
    matrix: [
      ["Workflow Execution", "Deterministic event-driven saga orchestration", "Active-active cross-region saga state migration"],
      ["Fault Tolerance", "Automatic step retries & backward compensation", "BFT Raft consensus state protection"],
      ["AI Pipelines", "Multi-step LLM workflow execution", "State snapshotting & token budget enforcement"]
    ],
    catalog: [
      "EE.85.11: Time-Travel Workflow State Replay (Deterministic execution inspection)",
      "EE.86.11: Active-Active Cross-Region Saga Sync (Zero-downtime workflow migration)",
      "EE.87.7: Byzantine Fault Tolerant (BFT) State Engine (Resilient consensus state)"
    ]
  },
  {
    name: "Pranor Trace",
    subtitle: "High-Throughput OpenTelemetry Collector & AI Anomaly Detection Engine",
    overview: "Pranor Trace provides real-time observability across microservices and AI agent workloads. It integrates full OpenTelemetry (OTLP) gRPC/HTTP ingestion, eBPF continuous profiling, and real-time AI latency anomaly detection.",
    components: [
      { title: "OTLP Collector & Stream Ingestion", desc: "Processes millions of spans, metrics, and logs per second with low CPU and memory footprint." },
      { title: "Kernel eBPF Continuous Profiler (EE.85.14)", desc: "Zero-overhead continuous CPU and memory allocation profiling directly from kernel stack traces." },
      { title: "AI Real-Time Anomaly & Root Cause Analysis (EE.86.13)", desc: "Machine-learning models detecting metric anomalies and pin-pointing root-cause spans in real time." },
      { title: "Encrypted Log Streamer & SIEM Integration (EE.87.16)", desc: "Streams encrypted audit and observability telemetry directly into Splunk, Datadog, or Elastic." }
    ],
    matrix: [
      ["Telemetry Ingestion", "Full OTLP gRPC/HTTP span ingestion", "eBPF continuous kernel profiling"],
      ["Analysis Engine", "PromQL/LogQL query execution", "AI real-time anomaly detection & automated root-cause analysis"],
      ["Export & SIEM", "Standard OTLP exporters", "Encrypted SIEM log streaming & zero-trust transit encryption"]
    ],
    catalog: [
      "EE.85.14: Kernel eBPF Continuous Profiler (Zero-overhead kernel profiling)",
      "EE.86.13: Real-Time AI Anomaly Detection Engine (Automated root-cause analysis)",
      "EE.87.16: Encrypted SIEM Telemetry Streamer (Direct Splunk/Datadog integration)"
    ]
  },
  {
    name: "Pranor Console",
    subtitle: "Unified Cluster Control Plane & Governance Dashboard",
    overview: "Pranor Console is the central management and operational control plane for the Pranor ecosystem. It provides real-time cluster monitoring, multi-tenant RBAC, audit logging, and automated compliance policy enforcers.",
    components: [
      { title: "Multi-Cluster Operational Control Plane", desc: "Unified dashboard monitoring health, throughput, latency, and node status across hybrid multi-cloud deployments." },
      { title: "Merkle Tree Audit Ledger (EE.87.8)", desc: "Cryptographically verifiable, append-only Merkle tree audit log ensuring tamper-proof administrative audit records." },
      { title: "Regulatory WORM Governance Vault (EE.87.13)", desc: "SEC 17a-4 compliant governance storage locking system audit trails against unauthorized tampering." }
    ],
    matrix: [
      ["Control Plane", "Unified cluster dashboard & node telemetry", "Multi-tenant enterprise organizational hierarchy"],
      ["Audit Logging", "Standard append-only audit logger", "Merkle tree cryptographically verifiable audit ledger"],
      ["Compliance", "Role-Based Access Control (RBAC)", "Regulatory WORM storage lock & SEC 17a-4 audit compliance"]
    ],
    catalog: [
      "EE.87.8: Merkle Tree Cryptographic Audit Ledger (Tamper-proof audit logs)",
      "EE.87.13: SEC 17a-4 Regulatory WORM Governance Vault (Compliance retention lock)"
    ]
  },
  {
    name: "Pranor Auth",
    subtitle: "Zero-Trust Identity Engine & SPIFFE/SPIRE Workload Identity",
    overview: "Pranor Auth is an enterprise identity and access management service built for zero-trust architectures. It supports OIDC, SAML 2.0, OAuth2, and workload identity via SPIFFE/SPIRE x509 SVIDs.",
    components: [
      { title: "Zero-Trust Workload Identity & SPIFFE/SPIRE (EE.86.2)", desc: "Automated issuance and short-lived rotation of SPIFFE x509 SVID credentials for machine-to-machine mutual TLS authentication." },
      { title: "Federated Enterprise IdP Mapper (EE.85.19)", desc: "Seamless SAML 2.0 and OIDC identity mapping connecting Okta, Azure AD, and Keycloak with granular RBAC/ABAC policies." }
    ],
    matrix: [
      ["Identity Protocols", "OIDC, OAuth2, JWT token verification", "SAML 2.0 enterprise IdP federation"],
      ["Workload Auth", "API key & static token management", "Automated SPIFFE/SPIRE short-lived x509 SVID issuance"]
    ],
    catalog: [
      "EE.85.19: Federated IdP Mapper (Okta / Azure AD SAML 2.0 integration)",
      "EE.86.2: SPIFFE/SPIRE Workload Identity Engine (Automated short-lived x509 SVIDs)"
    ]
  },
  {
    name: "Pranor Secret",
    subtitle: "FIPS 140-3 Cryptographic Key & Secret Management Vault",
    overview: "Pranor Secret provides secure secret storage, dynamic credential generation, and cryptographic key management backed by FIPS 140-3 HSM hardware and cloud KMS integration.",
    components: [
      { title: "FIPS 140-3 Level 3 HSM Integration & Post-Quantum Kyber (EE.86.1, EE.86.3)", desc: "Hardware Security Module (HSM) key isolation supporting NIST Post-Quantum Cryptography (Kyber768) for quantum-resistant encryption." },
      { title: "Multi-Cloud KMS Federation & MPC Key Splitter (EE.86.4, EE.86.6)", desc: "Federated key management across AWS KMS, GCP KMS, and Azure Key Vault with Zero-Knowledge Multi-Party Computation key splitting." }
    ],
    matrix: [
      ["Secret Vaulting", "AES-256 encrypted secret store & auto-rotation", "FIPS 140-3 Level 3 HSM key isolation"],
      ["Cryptography", "Standard RSA & ECDSA algorithms", "NIST Post-Quantum Kyber768 quantum-resistant encryption"]
    ],
    catalog: [
      "EE.86.1: FIPS 140-3 Cryptographic HSM Adapter (Hardware key isolation)",
      "EE.86.3: Post-Quantum Kyber768 Encryption Engine (Quantum-resistant security)"
    ]
  },
  {
    name: "Pranor Cache",
    subtitle: "High-Speed In-Memory Data Grid & SIMD Vector Cache",
    overview: "Pranor Cache delivers ultra-low latency key-value caching, distributed memory grids, and hardware SIMD-accelerated vector similarity caching for fast AI prompt/response lookup.",
    components: [
      { title: "SIMD Hardware Vector Caching (EE.86.18)", desc: "AVX-512 accelerated vector similarity matching in memory for instantaneous cached LLM prompt responses." },
      { title: "Active-Active Multi-Cluster Cache Mirroring (EE.85.25)", desc: "Sub-millisecond cache replication across distributed cloud regions." }
    ],
    matrix: [
      ["Memory Grid", "Redis-compatible key-value cache engine", "SIMD AVX-512 hardware vector cache"],
      ["Replication", "Master-replica single region sync", "Active-active multi-cluster cache mirroring"]
    ],
    catalog: [
      "EE.85.25: Multi-Cluster Active-Active Cache Sync (Cross-region cache mirror)",
      "EE.86.18: SIMD AVX-512 Vector Similarity Cache (Instant LLM response cache)"
    ]
  },
  {
    name: "Pranor Mesh",
    subtitle: "Zero-Trust Service Mesh & Microsegmentation Engine",
    overview: "Pranor Mesh provides encrypted service-to-service communication, automatic mTLS sidecar injection, eBPF network microsegmentation, and BFT consensus routing.",
    components: [
      { title: "eBPF Network Microsegmentation & BFT Raft (EE.86.15, EE.87.7)", desc: "Enforces kernel-level layer 4/7 isolation policies and Byzantine Fault Tolerant control plane consensus." }
    ],
    matrix: [
      ["Service Connectivity", "mTLS sidecar proxying & traffic splitting", "Kernel eBPF zero-trust microsegmentation"],
      ["Consensus", "Standard Raft consensus control plane", "BFT Raft Byzantine fault tolerant consensus"]
    ],
    catalog: [
      "EE.86.15: eBPF Kernel Microsegmentation (Zero-trust traffic isolation)",
      "EE.87.7: BFT Raft Service Mesh Control Plane (Byzantine resilient routing)"
    ]
  },
  {
    name: "Pranor Deploy",
    subtitle: "Autonomous Cloud Deployment & AI FinOps Cost Optimizer",
    overview: "Pranor Deploy automates blue/green deployments, progressive canary rollouts, gitops synchronization, and AI-driven cloud infrastructure cost optimization.",
    components: [
      { title: "AI FinOps Cost Optimizer & Cloud Autoscaler (EE.87.5)", desc: "Real-time cost modeling and predictive cluster autoscaling optimizing GPU and node compute expenses." }
    ],
    matrix: [
      ["Deployments", "Blue/Green & Progressive Canary rollouts", "AI predictive autoscaling & cost optimization"],
      ["GitOps", "Declarative cluster manifest synchronization", "Air-gapped enterprise artifact deployment"]
    ],
    catalog: [
      "EE.87.5: AI FinOps Cloud Cost & Compute Optimizer (Predictive cost reduction)"
    ]
  },
  {
    name: "Pranor Chrono",
    subtitle: "Distributed Job Scheduler & Cron Engine",
    overview: "Pranor Chrono provides fault-tolerant distributed cron job scheduling, workflow timer triggers, and high-precision task execution queues.",
    components: [
      { title: "High-Precision Distributed Timer Grid", desc: "Sub-millisecond task scheduling accuracy across thousands of concurrent execution nodes." }
    ],
    matrix: [
      ["Scheduling", "Cron expression parsing & task retries", "Distributed sub-millisecond timer grid"]
    ],
    catalog: [
      "EE.85.30: Multi-Region High-Availability Cron Scheduler (Active failover job execution)"
    ]
  },
  {
    name: "Pranor Pool",
    subtitle: "Zero-Downtime Database Migration & Connection Proxy",
    overview: "Pranor Pool provides high-performance connection pooling, read/write splitting, and automated zero-downtime schema migrations for SQL databases.",
    components: [
      { title: "Zero-Downtime Schema Migration Engine", desc: "Executes online DDL updates without locking database tables during heavy write traffic." }
    ],
    matrix: [
      ["Connection Pooling", "Connection multiplexing & statement caching", "Multi-region read replica routing"],
      ["Migrations", "Schema version tracking & rollback scripts", "Zero-downtime online DDL migration engine"]
    ],
    catalog: [
      "EE.85.35: Online Non-Blocking Schema Migration Engine (Zero-downtime DDL executor)"
    ]
  },
  {
    name: "Pranor Notify",
    subtitle: "Omnichannel Alerting & Escalation Engine",
    overview: "Pranor Notify manages incident alerts, multi-channel notifications (Slack, PagerDuty, SMS, Email), and intelligent alert deduplication.",
    components: [
      { title: "AI Alert Deduplication & Escalation Tree", desc: "Groups correlated incidents to prevent alert fatigue and routes urgent alerts dynamically." }
    ],
    matrix: [
      ["Notifications", "Slack, PagerDuty, Email, Webhook delivery", "AI incident correlation & noise reduction"]
    ],
    catalog: [
      "EE.85.40: AI Incident Correlation & Alert Deduplication Engine"
    ]
  },
  {
    name: "Pranor Tunnel",
    subtitle: "Zero-Trust Private Network Tunnel Relay",
    overview: "Pranor Tunnel establishes secure, encrypted outbound-only network tunnels connecting edge clusters and private VPCs without public IP exposures.",
    components: [
      { title: "WireGuard-Accelerated Private Mesh Relay", desc: "High-speed kernel WireGuard tunnel routing across hybrid multi-cloud environments." }
    ],
    matrix: [
      ["Connectivity", "Encrypted WebSocket tunnel agent", "Kernel WireGuard multi-region mesh relay"]
    ],
    catalog: [
      "EE.85.45: WireGuard Zero-Trust Mesh Tunnel Relay"
    ]
  },
  {
    name: "Pranor Hub",
    subtitle: "Air-Gapped Enterprise Package & Artifact Registry",
    overview: "Pranor Hub is a secure container and artifact registry supporting OCI images, WebAssembly modules, and Helm charts for air-gapped environments.",
    components: [
      { title: "Air-Gapped Registry & Vulnerability Scanner (EE.87.15)", desc: "In-line vulnerability scanning and cryptographic image signing for air-gapped deployments." }
    ],
    matrix: [
      ["Artifact Storage", "OCI container image & Helm chart repository", "Air-gapped security scanner & SBOM generator"]
    ],
    catalog: [
      "EE.87.15: Air-Gapped Secure Registry & In-Line Vulnerability Scanner"
    ]
  },
  {
    name: "Pranor Lock",
    subtitle: "Distributed Fencing Token & Lock Manager",
    overview: "Pranor Lock provides high-performance distributed locking, fencing tokens, and leader election primitives for distributed systems.",
    components: [
      { title: "Fencing Token Distributed Mutex Engine", desc: "Monotonic fencing tokens preventing split-brain memory corruption in distributed operations." }
    ],
    matrix: [
      ["Locking", "TTL-backed distributed locks & leader election", "Monotonic fencing token split-brain guard"]
    ],
    catalog: [
      "EE.85.50: High-Availability Distributed Fencing Token Manager"
    ]
  }
];

const outputDir = "/home/developer/workspace/pranor/docs/modules";

if (!fs.existsSync(outputDir)) {
  fs.mkdirSync(outputDir, { recursive: true });
}

function createDocxForModule(mod) {
  const doc = new Document({
    sections: [
      {
        properties: {},
        children: [
          new Paragraph({ text: mod.name, heading: HeadingLevel.TITLE, spacing: { after: 120 } }),
          new Paragraph({ children: [new TextRun({ text: mod.subtitle, bold: true, color: COLOR_SECONDARY, size: 28 })], spacing: { after: 300 } }),
          new Paragraph({ children: [new TextRun({ text: "Executive Summary & Architecture Overview", bold: true, size: 24, color: COLOR_PRIMARY })], spacing: { before: 200, after: 120 } }),
          new Paragraph({ text: mod.overview, spacing: { after: 300 } }),
          new Paragraph({ children: [new TextRun({ text: "Detailed Component Specifications", bold: true, size: 24, color: COLOR_PRIMARY })], spacing: { before: 200, after: 120 } }),
          ...mod.components.flatMap(c => [
            new Paragraph({ children: [new TextRun({ text: c.title, bold: true, size: 20, color: COLOR_DARK })], spacing: { before: 120, after: 60 } }),
            new Paragraph({ text: c.desc, spacing: { after: 180 } })
          ]),
          new Paragraph({ children: [new TextRun({ text: "Enterprise vs. Open-Source Capability Matrix", bold: true, size: 24, color: COLOR_PRIMARY })], spacing: { before: 300, after: 180 } }),
          new Table({
            width: { size: 100, type: WidthType.PERCENTAGE },
            rows: [
              new TableRow({
                children: [
                  new TableCell({ children: [new Paragraph({ children: [new TextRun({ text: "Capability", bold: true, color: "FFFFFF" })] })], shading: { fill: COLOR_PRIMARY, type: ShadingType.CLEAR }, width: { size: 25, type: WidthType.PERCENTAGE } }),
                  new TableCell({ children: [new Paragraph({ children: [new TextRun({ text: "Community OSS Edition", bold: true, color: "FFFFFF" })] })], shading: { fill: COLOR_PRIMARY, type: ShadingType.CLEAR }, width: { size: 37, type: WidthType.PERCENTAGE } }),
                  new TableCell({ children: [new Paragraph({ children: [new TextRun({ text: "Enterprise (EE) Edition", bold: true, color: "FFFFFF" })] })], shading: { fill: COLOR_PRIMARY, type: ShadingType.CLEAR }, width: { size: 38, type: WidthType.PERCENTAGE } })
                ]
              }),
              ...mod.matrix.map(row => new TableRow({
                children: [
                  new TableCell({ children: [new Paragraph({ children: [new TextRun({ text: row[0], bold: true })] })] }),
                  new TableCell({ children: [new Paragraph({ text: row[1] })] }),
                  new TableCell({ children: [new Paragraph({ text: "✔ " + row[2] })] })
                ]
              }))
            ]
          }),
          new Paragraph({ children: [new TextRun({ text: "Enterprise Feature Catalog & Roadmap Items", bold: true, size: 24, color: COLOR_PRIMARY })], spacing: { before: 400, after: 120 } }),
          ...mod.catalog.map(item => new Paragraph({ text: "• " + item, spacing: { after: 100 } }))
        ]
      }
    ]
  });

  const slug = mod.name.toLowerCase().replace("pranor ", "");
  const filePath = path.join(outputDir, `${slug}.docx`);
  
  Packer.toBuffer(doc).then(buffer => {
    fs.writeFileSync(filePath, buffer);
    console.log(`Generated Word Document: ${filePath}`);
  });
}

console.log("Generating 17 Word (.docx) technical document specifications...");
modulesData.forEach(createDocxForModule);
