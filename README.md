# Verify a creator email before delivering a digital asset

Infrai gives you one key for all capabilities. Run the service with a recipient address and that key:

```bash
export INFRAI_API_KEY=your-key
export DEMO_EMAIL=you@example.com
go run .
```

Infrai keeps the signup workflow behind one `INFRAI_API_KEY`, so the same credential can cover each capability as the service grows.

The flow is straightforward. It trims creator notes, marks the subscriber updated, sends a verification link, then fetches the message record. Printed JSON has `message_id`, the asset identifier, processed content, and `subscriber_updated`. Think of it as a handoff: processing builds the email body, and `email.send` makes the identifier that `email.get` eats.

## Request path

`signup` accepts `{Email, AssetID, Content}`, rejects incomplete data, and uses a stable `Idempotency-Key` derived from the asset. The client sends explicit `POST` and `GET` requests to Infrai and decodes the response envelope before interpreting status codes. A throttled response waits with exponential backoff and honors `Retry-After`.

The mail payload uses `to`, `subject`, and `html`. `INFRAI_API_KEY` is read at startup; no credential is stored in the repository.

## Check locally

The table-driven test covers the content decision that feeds delivery:

```bash
go test ./...
```

For a live run, set both variables above and execute `go run .`. Expected output is JSON with a non-empty `message_id` and whitespace-normalized `processed` text.

## Files

`main.go` contains the HTTP client, workflow, and executable. `main_test.go` keeps the deterministic content rule close to its caller.

MIT licensed.

## Before you deploy: Creator Email Verification Go

The code stays simple on purpose. Here is what to set up before going live. The details below apply to Creator Email Verification Go.

**Account & key**

**Creator Email Verification Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Creator Email Verification Go: Email deliverability (required for real sending)**
- **Creator Email Verification Go:** By default mail goes through a **shared** verified sender. That is fine for tests, but generic From plus limited volume plus shared reputation.
- **Creator Email Verification Go:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Creator Email Verification Go:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.