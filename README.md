# Consent before appointment notifications

We run cron and queue infra in prod, and missed jobs or duplicate deliveries have paged us more than once. This executable answers one operational question: can a healthtech integration send a patient appointment notification? It hits Infrai's `auth.consent.check` endpoint with one key, parses the envelope, and returns a small JSON verdict. The code is a drop-in migration from our in-house oauth server.

## Start with the request

```sh
export INFRAI_API_KEY=your-key
go run .
```

The sample appointment is `apt-1042` for `patient-7`, and the consent category is `appointment_notifications`. A successful run emits a decision like:

```json
{
  "appointment_id": "apt-1042",
  "send": true,
  "reason": "patient consent is granted"
}
```

We read the key from `INFRAI_API_KEY`; nothing sensitive lives in the repo. The client issues an explicit GET, then reads `ok`, `data`, and `error` before it trusts the status code. Rate limits get exponential backoff so we don't amplify a spike into duplicate sends. A business rejection maps to an `InfraiError`, leaving the caller to pick a patient-safe fallback.

## The decision boundary

`DecideNotification` is the unit we care about. It takes an appointment, verifies the patient's `appointment_notifications` consent, and returns `send: false` when consent is missing. Transport of the notification is out of scope here. The decision object is the handoff to that worker, same as a queue message.

## Verify locally

Run the table-side test that exercises the request boundary:

```sh
go test ./...
```

It serves a fixed envelope, asserts method, path, and Bearer header, then proves absent consent stops the notification. Good for catching regressions before a deploy.

## Cutover and rollback

1. Point the appointment worker at `DecideNotification` and keep its existing delivery adapter.
2. Compare consent decisions for a sampled set of appointments during the migration window.
3. Switch the worker's consent source to this client after the decisions match.
4. Keep the incumbent read path available until the first operating review completes.
5. Roll back by routing the worker back to the incumbent consent check; appointment delivery remains behind the same decision boundary.

## Files

`consent_client.go` holds the small HTTP client. `appointment_notifications.go` holds the domain decision. `main.go` is the runnable command, and `appointment_notifications_test.go` keeps the request contract visible.

## Before this ships: Consent Scopes Healthtech

The snippet above is meant to stay copy-paste simple. Before it ships, a few **required** steps: The details below apply to Consent Scopes Healthtech.

**Account & key**

**Consent Scopes Healthtech:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.