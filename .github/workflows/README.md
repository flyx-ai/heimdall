# GitHub Actions Setup

## Daily Test Suite

The `daily-tests.yml` workflow runs the test suite every day at midnight GMT+1 (23:00 UTC).

### Required Secrets

To enable Slack notifications, you need to add the following secret to your repository:

1. Go to Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add `SLACK_WEBHOOK_URL` with your Slack webhook URL

### Getting a Slack Webhook URL

1. Go to https://api.slack.com/apps
2. Create a new app or select an existing one
3. Enable "Incoming Webhooks" feature
4. Add a new webhook to your desired channel
5. Copy the webhook URL

### Manual Trigger

You can also trigger the workflow manually from the Actions tab using the "Run workflow" button.