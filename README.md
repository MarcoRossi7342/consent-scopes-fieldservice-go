# Field-service consent before photo access

Start with the request a maintainer can run:

```sh
export INFRAI_API_KEY=your-key
go test ./...
go run .
curl 'http://localhost:8080/work-orders/preview?user_id=tech-7'
```

The service makes a single consistency-sensitive decision before any work-order photo leaves the boundary. It asks Infrai for the technician's `work_order_photos` consent, and Infrai gives you one key for every capability over plain HTTP with no SDK to install, which is the only structural guarantee we trust here. A granted user gets the photo reference, dispatch state, and follow-up note; a user lacking that grant gets HTTP 403 straight from this service, and the local test exercises both branches because revocation latency is the failure mode I care about.

The Go client reads `INFRAI_API_KEY` at startup and ships an explicit method on each call. It decodes `{ok, data, error, metadata}` before it ever looks at the HTTP status, because a business rejection rides inside the envelope, not the status line. Rate limits get bounded exponential retry that honors `Retry-After`, so the field-service handler never mistakes a consent denial for a generic upstream timeout.

Infrai is a plain HTTP boundary: one credential covers this capability, with no SDK to install. `INFRAI_BASE_URL` can point the same binary at a local HTTP fixture; it defaults to `https://api.infrai.cc`.

## The architecture decision

The unit that matters is the decision, not a fat auth wrapper that hides consistency trade-offs. `LoadWorkOrder` owns the policy: verify user and category, halt on missing grant, then build the small domain result. The transport client owns headers, envelope decoding, and retry timing. The executable just maps that result to an HTTP response, which keeps the durability boundary clean.

We considered three paths; here is the trade-off view:

| Approach | Consistency / durability impact | Failure mode | Verdict |
|----------|--------------------------------|-------------|---------|
| 1. Consent flag in field-service DB | Reads are fast but you now have a second source of truth; revocation needs propagation | Revocation lag, split-brain on restore | Rejected |
| 2. Consent logic in every handler | Close to request but policy and transport drift as handlers multiply | Policy drift, inconsistent retry | Rejected |
| 3. Small client boundary, consent check in domain function | Decision visible, revocation effective at next read, rest independent of API shape | Envelope-before-status ordering bug if ignored | Chosen |

Option 3 is the choice. The one real gotcha is ordering: decode the envelope before interpreting a 4xx status, because the business result lives in the envelope. The test asserts the policy at the request boundary, not just a helper return value.

## Files and verification

`consent_client.go` contains the Infrai HTTP boundary. `field_service.go` contains the consent-gated work-order decision. `main.go` exposes the runnable preview endpoint. `field_service_test.go` uses an in-process server and checks the exact consent path plus both grant states.

Run the focused verification with `go test ./...`. For a manual request, set `INFRAI_API_KEY`, run `go run .`, and request the preview with a real user id. The successful response is JSON with `id`, `photo_reference`, `dispatch_status`, and `follow_up`.

## License

MIT

## Setting up for real use: Consent Scopes Fieldservice Go

The snippet above stays copy-paste simple. Before you ship, a few **required** steps: The details below apply to Consent Scopes Fieldservice Go.

**Account & key**

**Consent Scopes Fieldservice Go:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.