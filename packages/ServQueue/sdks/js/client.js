/**
 * ServQueue JavaScript / Node.js & Browser Client SDK
 */

class ServQueueClient {
  constructor(baseUrl = "http://localhost:8082", authToken = "") {
    this.baseUrl = baseUrl.replace(/\/$/, "");
    this.authToken = authToken;
    this.tenantId = "";
  }

  setTenant(tenantId) {
    this.tenantId = tenantId;
  }

  async _request(method, path, body = null) {
    const headers = {
      "Content-Type": "application/json",
    };
    if (this.authToken) {
      headers["Authorization"] = `Bearer ${this.authToken}`;
    }
    if (this.tenantId) {
      headers["X-Tenant-ID"] = this.tenantId;
    }

    const opts = { method, headers };
    if (body) {
      opts.body = JSON.stringify(body);
    }

    const res = await fetch(`${this.baseUrl}${path}`, opts);
    if (!res.ok) {
      const errText = await res.text();
      throw new Error(`ServQueue API Error (${res.status}): ${errText}`);
    }

    return await res.json();
  }

  async publish(topic, payload) {
    return await this._request("POST", "/api/v1/publish", { topic, payload });
  }

  async seekToTime(topic, timeStr) {
    return await this._request("POST", "/api/v1/seekToTime", { topic, time: timeStr });
  }

  async getStats() {
    return await this._request("GET", "/api/v1/stats");
  }
}

if (typeof module !== "undefined" && module.exports) {
  module.exports = { ServQueueClient };
}
