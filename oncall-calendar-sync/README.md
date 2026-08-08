# Jira On-Call Calendar Sync

Syncs your **Jira Service Management (JSM) Schedules** on-call rotation into a shared **Google Calendar** via Google Apps Script.

- Reads the schedule **timeline** from the [JSM operations REST API](https://developer.atlassian.com/cloud/jira/service-desk-ops/rest/v2/intro/) (`https://api.atlassian.com/jsm/ops/api/{cloudId}/v1/...`), which includes base periods, overrides and forwardings (Basic auth with an Atlassian API token — works even when SAML SSO is enforced).
- Upserts events idempotently (tagged with `shiftId` in extended properties): creates new, updates changed, deletes removed shifts.
- Adds the **shift owner as an attendee** by default, so they get their own invite + notifications. Reassignments update the guest automatically.
- Runs on a time trigger (default every 6h) so shift changes propagate without manual imports.

## How it works

```
JSM ops API (schedule timeline) ──Basic auth (email + API token)──► Apps Script (time trigger)
                                                                        │
                                                                        ▼
                                                         Google Calendar "On-Call"
                                                          - events per shift
                                                          - shift owner as guest
                                                          - shared read-only with team
```

The script auto-discovers your `cloudId` from `https://<site>._edge/tenant_info` and caches it, so no extra setup is needed beyond `JIRA_SITE`.

## 1. Prerequisites (one-time, ~5 min)

### a) Atlassian API token

1. Go to <https://id.atlassian.com/manage-profile/security/api-tokens>
2. Click **Create API token** → name it `oncall-calendar-sync` → copy it.
3. Note your account **email** (the one tied to the token) and your **Jira site URL** (e.g. `https://yourcompany.atlassian.net`).

> SSO note: even if your org enforces SAML SSO for browser logins, API tokens work for REST API automation.

### b) Find the schedule ID

Either grab it from the JSM schedule's URL in the UI, or after the script is deployed run `listSchedules()` and read the ID from the execution log.

## 2. Deploy the script

### Option A — clasp (CLI, recommended)

```bash
npm install -g @google/clasp
cd ~/dev/projects/study-topics/oncall-calendar-sync
clasp login            # opens browser, pick your Google account
clasp create --type standalone --title "Jira On-Call Calendar Sync"
clasp push             # uploads appsscript.json + src/SyncOnCall.gs
```

The `appsscript.json` manifest already enables the **Calendar (v3) advanced service** and the required OAuth scopes — no manual enabling needed.

### Option B — Manual (no CLI)

1. Go to <https://script.google.com> → **New project** → name it "Jira On-Call Calendar Sync".
2. Paste the contents of `src/SyncOnCall.gs` into `Code.gs`.
3. Open **Project Settings** → check the manifest shows the advanced service enabled, or open the **Editor** → `+` next to Services → add **Google Calendar API** → enable.
4. Open **Project Settings** → check **Time zone** is correct.

## 3. Configure

Run this once in the editor (modify the values):

```js
setupConfig({
  JIRA_SITE: 'https://yourcompany.atlassian.net',
  JIRA_EMAIL: 'you@company.com',
  JIRA_API_TOKEN: 'the-token-from-step-1',
  JIRA_SCHEDULE_ID: 'the-schedule-id',
  // optional overrides:
  SKIP_ROTATION_NAMES: 'Working Hours',   // comma-separated substrings to skip (case-insensitive contains)
  CALENDAR_NAME: 'On-Call',      // default
  ADD_GUESTS: 'true',            // default: invite shift owner as attendee
  EVENT_REMINDER_MINUTES: '0',   // 0 = calendar default; e.g. 60 = popup 1h before
  TRIGGER_HOURS: '6',            // resync interval
});
```

Alternative: **Project Settings → Script properties** and add the keys above (`JIRA_API_TOKEN` etc.) manually.

If you don't know the schedule ID, run `listSchedules()` first (after setting site/email/token) and read the log. If cloud ID auto-discovery fails (unusual), set `CLOUD_ID` explicitly — you can find it in the Jira site URL of the admin UI or via `GET https://<site>.atlassian.net/_edge/tenant_info`.

## 4. Run it

From the editor, in order:

1. `dryRun()` — logs exactly what would be created/updated/deleted, changes nothing. Check **Executions → log**.
2. `syncCalendar()` — performs the sync.
3. `installTrigger()` — schedules `syncCalendar` to run every `TRIGGER_HOURS` hours.
4. (Optional) `runAll()` — does all of the above in one call.

Open <https://calendar.google.com> → you should now see the **"On-Call"** calendar with your shifts.

## 5. Share with the team

1. <https://calendar.google.com> → find the **On-Call** calendar → **Settings and sharing**.
2. Under **Share with specific people or groups**, add your teammates **View all event details** (read-only is enough).
3. They'll see all shifts; the shift owner additionally receives an invite (attendee) with their own notifications.

Tips:

- Reminders: if you set `EVENT_REMINDER_MINUTES`, it applies to everyone viewing the calendar. Otherwise each viewer's own calendar settings decide. Note Google Calendar has no "notify me for events matching pattern X" — it's per-calendar or per-event.
- Guests on by default: `ADD_GUESTS='true'` adds the shift owner as an attendee, so the invite + reminders reach them even if they never open the shared calendar. Flip to `'false'` if invite emails get noisy.

## Troubleshooting

| Problem | Fix |
|---|---|
| `JSM ops API auth failed (401)` | Wrong `JIRA_EMAIL` / `JIRA_API_TOKEN`. Regenerate the token at id.atlassian.com. |
| `JSM ops API forbidden (403)` | Your user lacks access to the schedule: must be a rotation responder, a member of the schedule's team, or have read-only admin on it. |
| `JSM ops API 404` | Wrong `JIRA_SCHEDULE_ID` (run `listSchedules()` to find the right one) or wrong `CLOUD_ID`. |
| `JSM ops API rate limited (429)` | Too many calls in a row — wait and re-run. |
| `Missing config` | Run `setupConfig({...})` with all required fields. |
| `Events not updating` | Re-run `syncCalendar()`; check the trigger is installed (`installTrigger()`). |
| Re-run creates duplicates | Shouldn't happen (idempotent by `shiftId`). If it ever does: `deleteAllSyncedEvents()` then `syncCalendar()`. |
| Guest not added | The responder's email couldn't be resolved (Jira user API may not return emails for all users; team/escalation responders have no attendee). Log shows the reason. |
| Shifts are missing | The timeline window is today → 14 days ahead. If you need further-out shifts, adjust the `interval`/`intervalUnit` params in `fetchAllShifts()` (timeline API allows up to 2 years forward). |
| Some rotations appear that shouldn't | Run `showTimeline()` to see the raw rotation names, then set `SKIP_ROTATION_NAMES` to skip them (e.g. `'Working Hours'`). |

## Security

- The API token lives in Apps Script **Script Properties** (encrypted at rest, visible only to editors of the script project). Never hardcode it in code or commit it.
- Keep the script project shared only with people you trust — it can read your Jira schedule.

## Files

```
oncall-calendar-sync/
├── README.md           # this file
├── appsscript.json     # manifest: advanced Calendar service + OAuth scopes
└── src/
    └── SyncOnCall.gs   # the whole script (config, Jira fetch, calendar sync, triggers)
```
