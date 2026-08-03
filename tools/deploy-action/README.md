# Pranor Deploy GitHub Action

This GitHub Action builds a Pranor `.pnr` application, pushes it to **Pranor Hub**, and initiates a zero-downtime blue/green deployment via **Pranor Deploy**.

## Usage

```yaml
uses: vyuvaraj/pranor/tools/deploy-action@v1
with:
  app_path: '.'
  pranor_hub_url: 'http://hub.pranor.internal:8082'
  pranor_deploy_url: 'http://deploy.pranor.internal:8088'
  api_token: ${{ secrets.PRANOR_API_TOKEN }}
```
