"""
PranorPulse Python Client SDK
"""

import json
import urllib.request
import urllib.error

class PranorPulseClient:
    def __init__(self, base_url="http://localhost:8082", auth_token=""):
        self.base_url = base_url.rstrip("/")
        self.auth_token = auth_token
        self.tenant_id = ""

    def set_tenant(self, tenant_id):
        self.tenant_id = tenant_id

    def _request(self, method, path, body=None):
        url = f"{self.base_url}{path}"
        headers = {"Content-Type": "application/json"}
        if self.auth_token:
            headers["Authorization"] = f"Bearer {self.auth_token}"
        if self.tenant_id:
            headers["X-Tenant-ID"] = self.tenant_id

        data = json.dumps(body).encode("utf-8") if body else None
        req = urllib.request.Request(url, data=data, headers=headers, method=method)

        try:
            with urllib.request.urlopen(req) as resp:
                res_body = resp.read().decode("utf-8")
                return json.loads(res_body) if res_body else {}
        except urllib.error.HTTPError as e:
            err_body = e.read().decode("utf-8")
            raise RuntimeError(f"PranorPulse API Error ({e.code}): {err_body}")

    def publish(self, topic: str, payload: str):
        return self._request("POST", "/api/v1/publish", {"topic": topic, "payload": payload})

    def seek_to_time(self, topic: str, time_str: str):
        return self._request("POST", "/api/v1/seekToTime", {"topic": topic, "time": time_str})

    def get_stats(self):
        return self._request("GET", "/api/v1/stats")
