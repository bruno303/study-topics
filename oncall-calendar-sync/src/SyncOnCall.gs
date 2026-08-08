/**
 * Jira Service Management (JSM) Schedules -> Google Calendar sync.
 *
 * Reads on-call shifts from the JSM operations REST API (schedule timeline)
 * and upserts them as events on a shared Google Calendar. Each shift is
 * tagged with an extendedProperty (shiftId) so re-runs are idempotent:
 * create new, update changed, delete removed. The shift owner is added as
 * an attendee so they get their own invite + notifications.
 *
 * API: https://api.atlassian.com/jsm/ops/api/{cloudId}/v1/...
 * Auth: Atlassian API token (id.atlassian.com) with Basic auth. Works even
 * when the org enforces SAML SSO for browser logins.
 */

const CONFIG_KEYS = {
  JIRA_SITE: 'JIRA_SITE',                   // e.g. https://yourcompany.atlassian.net (no trailing slash)
  JIRA_EMAIL: 'JIRA_EMAIL',                 // account email used to create the API token
  JIRA_API_TOKEN: 'JIRA_API_TOKEN',
  JIRA_SCHEDULE_ID: 'JIRA_SCHEDULE_ID',     // leave empty -> run listSchedules() to discover
  CLOUD_ID: 'CLOUD_ID',                     // optional; auto-discovered from JIRA_SITE when unset
  SKIP_ROTATION_NAMES: 'SKIP_ROTATION_NAMES', // comma-separated substrings to skip (case-insensitive contains)
  CALENDAR_NAME: 'CALENDAR_NAME',           // default: 'On-Call'
  ADD_GUESTS: 'ADD_GUESTS',                 // 'true' | 'false' (default true): invite shift owner as attendee
  EVENT_REMINDER_MINUTES: 'EVENT_REMINDER_MINUTES', // 0 = calendar default; else popup N minutes before start
  TRIGGER_HOURS: 'TRIGGER_HOURS',           // resync interval, default 6
};

const DEFAULTS = {
  CALENDAR_NAME: 'On-Call',
  ADD_GUESTS: true,
  EVENT_REMINDER_MINUTES: 0,
  TRIGGER_HOURS: 6,
};

const OPS_BASE = 'https://api.atlassian.com/jsm/ops/api/';

/* ------------------------------------------------------------------ */
/* Config                                                              */
/* ------------------------------------------------------------------ */

function getConfig() {
  const props = PropertiesService.getScriptProperties();
  const cfg = {};
  for (const key in CONFIG_KEYS) {
    cfg[key] = props.getProperty(CONFIG_KEYS[key]);
  }
  return {
    site: normalizeSite(cfg.JIRA_SITE),
    email: cfg.JIRA_EMAIL,
    apiToken: cfg.JIRA_API_TOKEN,
    scheduleId: cfg.JIRA_SCHEDULE_ID,
    cloudId: cfg.CLOUD_ID,
    skipRotationNames: (cfg.SKIP_ROTATION_NAMES || '').split(',').map(s => s.trim().toLowerCase()).filter(Boolean),
    calendarName: cfg.CALENDAR_NAME || DEFAULTS.CALENDAR_NAME,
    addGuests: (cfg.ADD_GUESTS === null || cfg.ADD_GUESTS === undefined || cfg.ADD_GUESTS === '')
      ? DEFAULTS.ADD_GUESTS : String(cfg.ADD_GUESTS).toLowerCase() === 'true',
    reminderMinutes: parseInt(cfg.EVENT_REMINDER_MINUTES, 10) || DEFAULTS.EVENT_REMINDER_MINUTES,
    triggerHours: parseInt(cfg.TRIGGER_HOURS, 10) || DEFAULTS.TRIGGER_HOURS,
  };
}

/** Ensure the site is a full https URL (e.g. "company.atlassian.net" -> "https://..."). */
function normalizeSite(site) {
  if (!site) return '';
  const s = site.replace(/\/+$/, '');
  return /^https?:\/\//i.test(s) ? s : 'https://' + s;
}

/**
 * One-time setup helper. Call from the editor, e.g.:
 *   setupConfig({ JIRA_SITE: 'https://yourcompany.atlassian.net',
 *                 JIRA_EMAIL: 'you@company.com',
 *                 JIRA_API_TOKEN: 'xxxx', JIRA_SCHEDULE_ID: 'yyy' })
 */
function setupConfig(values) {
  const props = PropertiesService.getScriptProperties();
  for (const key in CONFIG_KEYS) {
    if (values[CONFIG_KEYS[key]]) {
      props.setProperty(CONFIG_KEYS[key], values[CONFIG_KEYS[key]]);
    }
  }
  Logger.log('Config saved. Current config:');
  const cfg = getConfig();
  for (const key in CONFIG_KEYS) {
    const v = props.getProperty(CONFIG_KEYS[key]);
    const masked = CONFIG_KEYS[key] === 'JIRA_API_TOKEN' && v ? '<redacted>' : v;
    Logger.log('  %s = %s', CONFIG_KEYS[key], masked);
  }
  Logger.log('API token missing? JIRA_API_TOKEN unset -> set it via setupConfig({...}).');
}

function assertConfig(cfg) {
  const missing = [];
  if (!cfg.site) missing.push('JIRA_SITE');
  if (!cfg.email) missing.push('JIRA_EMAIL');
  if (!cfg.apiToken) missing.push('JIRA_API_TOKEN');
  if (missing.length) {
    throw new Error('Missing config: ' + missing.join(', ') +
      '. Run setupConfig({...}) first (see README).');
  }
}

/* ------------------------------------------------------------------ */
/* JSM ops API                                                         */
/* ------------------------------------------------------------------ */

function authHeader(cfg) {
  return 'Basic ' + Utilities.base64Encode(cfg.email + ':' + cfg.apiToken);
}

/**
 * Cloud ID of the Jira instance. Auto-discovered from the site and cached
 * in script properties; set CLOUD_ID explicitly to override.
 */
function getCloudId(cfg) {
  if (cfg.cloudId) return cfg.cloudId;
  const props = PropertiesService.getScriptProperties();
  const cached = props.getProperty('_CLOUD_ID');
  if (cached) return cached;
  const resp = UrlFetchApp.fetch(cfg.site + '/_edge/tenant_info', {
    method: 'get',
    headers: { Accept: 'application/json' },
    muteHttpExceptions: true,
  });
  const code = resp.getResponseCode();
  const body = resp.getContentText();
  if (code < 200 || code >= 300) {
    throw new Error('Could not discover cloudId from ' + cfg.site +
      ' (' + code + '). Set CLOUD_ID manually in setupConfig({...}).');
  }
  const info = JSON.parse(body);
  if (!info.cloudId) {
    throw new Error('tenant_info response had no cloudId. Set CLOUD_ID manually.');
  }
  props.setProperty('_CLOUD_ID', info.cloudId);
  Logger.log('Discovered cloudId: %s (cached)', info.cloudId);
  return info.cloudId;
}

/**
 * Authenticated GET against the JSM ops REST API
 * (https://api.atlassian.com/jsm/ops/api/{cloudId}/v1{path}).
 */
function opsFetch(cfg, path) {
  const cloudId = getCloudId(cfg);
  const url = OPS_BASE + encodeURIComponent(cloudId) + '/v1' + path;
  return httpGet(cfg, url, 'JSM ops API');
}

/** Authenticated GET against the Jira site API (e.g. /rest/api/3/user). */
function siteFetch(cfg, path) {
  return httpGet(cfg, cfg.site + path, 'Jira API');
}

function httpGet(cfg, url, label) {
  const resp = UrlFetchApp.fetch(url, {
    method: 'get',
    headers: { Authorization: authHeader(cfg), Accept: 'application/json' },
    muteHttpExceptions: true,
  });
  const code = resp.getResponseCode();
  const body = resp.getContentText();
  if (code < 200 || code >= 300) {
    let detail = '';
    try {
      const j = JSON.parse(body);
      detail = (j.errors || []).map(e => e.title).join(', ') || j.message || '';
    } catch (e) { /* keep empty */ }
    const suffix = detail ? ' (' + detail + ')' : '';
    if (code === 401) {
      throw new Error(label + ' auth failed (401) - check JIRA_EMAIL / JIRA_API_TOKEN.' + suffix);
    }
    if (code === 403) {
      throw new Error(label + ' forbidden (403) - your user lacks access; you must be a ' +
        'rotation responder, a member of the schedule team, or have read-only admin.' + suffix);
    }
    if (code === 404) {
      throw new Error(label + ' 404 - check JIRA_SCHEDULE_ID (run listSchedules()) and CLOUD_ID.' + suffix);
    }
    if (code === 429) {
      throw new Error(label + ' rate limited (429) - retry later.' + suffix);
    }
    throw new Error(label + ' failed (' + code + ') on ' + url + ': ' + (detail || body.slice(0, 300)));
  }
  return body ? JSON.parse(body) : null;
}

/**
 * Discovery helper: logs every schedule with its id + rotation summary.
 * Run this from the editor once config is set to find JIRA_SCHEDULE_ID.
 */
function listSchedules() {
  const cfg = getConfig();
  assertConfig(cfg);
  const schedules = [];
  let offset = 0;
  const size = 50;
  while (true) {
    const data = opsFetch(cfg, '/schedules?offset=' + offset + '&size=' + size);
    const page = data.values || [];
    schedules.push.apply(schedules, page);
    const next = data.links && data.links.next;
    if (!next || page.length < size || schedules.length > 1000) break;
    offset += page.length;
  }
  Logger.log('Found %s schedule(s):', schedules.length);
  for (const s of schedules) {
    Logger.log('  [%s] %s', s.id, s.name);
  }
  return schedules;
}

/** Debug helper: dumps the raw schedule timeline of the configured schedule. */
function showTimeline() {
  const cfg = getConfig();
  assertConfig(cfg);
  if (!cfg.scheduleId) throw new Error('JIRA_SCHEDULE_ID not set - run listSchedules() first.');
  const today = startOfTodayUTC();
  const data = opsFetch(cfg, '/schedules/' + encodeURIComponent(cfg.scheduleId) +
    '/timeline?date=' + isoDate(today) + '&interval=14&intervalUnit=days');
  Logger.log(JSON.stringify(data, null, 2));
  return data;
}

/**
 * Fetch all on-call shifts of the configured schedule from its timeline
 * (base periods + overrides + forwardings already merged into finalTimeline).
 * Window: today at 00:00 UTC → 14 days ahead.
 */
function fetchAllShifts(cfg) {
  if (!cfg.scheduleId) throw new Error('JIRA_SCHEDULE_ID not set - run listSchedules() first.');
  const today = startOfTodayUTC();
  const data = opsFetch(cfg, '/schedules/' + encodeURIComponent(cfg.scheduleId) +
    '/timeline?date=' + isoDate(today) + '&interval=14&intervalUnit=days');
  const rotations = (data.finalTimeline && data.finalTimeline.rotations) || [];
  const userCache = {};
  const shifts = [];
  for (const rot of rotations) {
    if (rot.deleted) continue;
    const rotName = (rot.name || '').toLowerCase();
    if (cfg.skipRotationNames.some(p => rotName.includes(p))) {
      Logger.log('Skipping rotation "%s" (matches skip pattern)', rot.name);
      continue;
    }
    for (const period of (rot.periods || [])) {
      if (!period.responder || period.responder.type === 'noone') continue;
      const start = new Date(period.startDate);
      const end = new Date(period.endDate);
      if (isNaN(start.getTime()) || isNaN(end.getTime())) {
        Logger.log('Skipping malformed period in rotation "%s": %s', rot.name || '', JSON.stringify(period));
        continue;
      }
      const resolved = resolveResponder(cfg, period, userCache);
      shifts.push({
        shiftId: cfg.scheduleId + '/' + rot.id + '/' + period.startDate,
        rotationName: rot.name || '',
        responderName: resolved.name || 'Unknown',
        email: resolved.email,
        periodType: period.type || 'base',
        start: start,
        end: end,
      });
    }
  }
  shifts.sort((a, b) => a.start - b.start);
  Logger.log('Fetched %s shift(s) from schedule timeline.', shifts.length);
  return shifts;
}

/**
 * Resolve a timeline period's responder into a display name (+ email for
 * attendees). User responders are looked up via the Jira user API (cached
 * per run). Team/escalation responders fall back to their flattened users.
 */
function resolveResponder(cfg, period, userCache) {
  const resp = period.responder || {};
  const type = resp.type || 'unknown';
  const id = resp.id;
  if (type === 'user') {
    const u = getUser(cfg, id, userCache);
    return {
      name: (u && u.displayName) || resp.displayName || 'User',
      email: cfg.addGuests && u ? (u.emailAddress || null) : null,
    };
  }
  const users = (period.flattenedResponders || []).filter(r => r.type === 'user');
  if (users.length === 1) {
    const u = getUser(cfg, users[0].id, userCache);
    return {
      name: (u && u.displayName) || 'Team member',
      email: cfg.addGuests && u ? (u.emailAddress || null) : null,
    };
  }
  return { name: type === 'team' ? 'Team' : 'Escalation', email: null };
}

/** Look up a Jira user by accountId (cached); returns null on failure. */
function getUser(cfg, accountId, cache) {
  if (cache[accountId] !== undefined) return cache[accountId];
  try {
    const u = siteFetch(cfg, '/rest/api/3/user?accountId=' + encodeURIComponent(accountId));
    cache[accountId] = u || null;
  } catch (e) {
    Logger.log('User lookup failed for %s: %s', accountId, e.message);
    cache[accountId] = null;
  }
  return cache[accountId];
}

/* ------------------------------------------------------------------ */
/* Calendar sync                                                       */
/* ------------------------------------------------------------------ */

function getOrCreateCalendar(cfg) {
  const existing = CalendarApp.getCalendarsByName(cfg.calendarName);
  if (existing.length > 0) return existing[0];
  Logger.log('Creating calendar "%s"', cfg.calendarName);
  const cal = CalendarApp.createCalendar(cfg.calendarName);
  cal.setDescription('On-call shifts synced from JSM Schedules (read-only source: Jira).');
  return cal;
}

function toIso(date) {
  return Utilities.formatDate(date, 'UTC', "yyyy-MM-dd'T'HH:mm:ss'Z'");
}

function isoDate(date) {
  const y = date.getUTCFullYear();
  const m = String(date.getUTCMonth() + 1).padStart(2, '0');
  const d = String(date.getUTCDate()).padStart(2, '0');
  const hh = String(date.getUTCHours()).padStart(2, '0');
  const mm = String(date.getUTCMinutes()).padStart(2, '0');
  const ss = String(date.getUTCSeconds()).padStart(2, '0');
  return encodeURIComponent(y + '-' + m + '-' + d + 'T' + hh + ':' + mm + ':' + ss + '+00:00');
}

/** Returns today at 00:00:00 UTC. */
function startOfTodayUTC() {
  const now = new Date();
  return new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate()));
}

function buildEventPayload(shift, cfg) {
  const attendees = shift.email ? [{ email: shift.email }] : [];
  return {
    summary: 'On-Call: ' + shift.responderName,
    description: 'Rotation: ' + (shift.rotationName || '-') + '\nPeriod: ' + shift.periodType +
      '\nSource: JSM Schedules\nshiftId: ' + shift.shiftId,
    start: { dateTime: toIso(shift.start), timeZone: 'UTC' },
    end: { dateTime: toIso(shift.end), timeZone: 'UTC' },
    attendees: attendees,
    guestsCanModify: false,
    guestsCanInviteOthers: false,
    extendedProperties: {
      private: { shiftId: shift.shiftId, rotation: shift.rotationName || '' },
    },
    reminders: cfg.reminderMinutes > 0
      ? { useDefault: false, overrides: [{ method: 'popup', minutes: cfg.reminderMinutes }] }
      : { useDefault: true },
  };
}

/** Map shiftId -> event for all events we manage on the calendar. */
function listSyncedEvents(calId) {
  const out = {};
  let pageToken = null;
  do {
    const res = Calendar.Events.list(calId, {
      privateExtendedProperty: 'shiftId=*',
      singleEvents: true,
      maxResults: 250,
      pageToken: pageToken,
      fields: 'nextPageToken,items(id,summary,start,end,attendees,extendedProperties,reminders)',
    });
    for (const item of (res.items || [])) {
      const shiftId = item.extendedProperties && item.extendedProperties.private
        ? item.extendedProperties.private.shiftId : null;
      if (shiftId) out[shiftId] = item;
    }
    pageToken = res.nextPageToken || null;
  } while (pageToken);
  return out;
}

function sameEvent(existing, payload) {
  const existingStart = existing.start && existing.start.dateTime;
  const existingEnd = existing.end && existing.end.dateTime;
  if (existing.summary !== payload.summary) return false;
  if (existingStart !== payload.start.dateTime || existingEnd !== payload.end.dateTime) return false;
  const oldEmails = (existing.attendees || []).map(a => a.email).sort().join(',');
  const newEmails = (payload.attendees || []).map(a => a.email).sort().join(',');
  if (oldEmails !== newEmails) return false;
  if (existing.description !== payload.description) return false;
  return true;
}

/**
 * Compute what would change without touching the calendar.
 * Returns { creates: [], updates: [], deletes: [] } with human-readable lines.
 */
function planChanges(cfg) {
  assertConfig(cfg);
  const shifts = fetchAllShifts(cfg);
  const cal = getOrCreateCalendar(cfg);
  const existing = listSyncedEvents(cal.getId());
  const plan = { creates: [], updates: [], deletes: [] };
  const seen = {};
  for (const s of shifts) {
    seen[s.shiftId] = true;
    const payload = buildEventPayload(s, cfg);
    const ev = existing[s.shiftId];
    const label = s.responderName + ' (' + toIso(s.start) + ' -> ' + toIso(s.end) + ')';
    if (!ev) {
      plan.creates.push(label + (payload.attendees.length ? ' [guest ' + payload.attendees[0].email + ']' : ''));
    } else if (!sameEvent(ev, payload)) {
      plan.updates.push(label);
    }
  }
  for (const key in existing) {
    if (!seen[key]) plan.deletes.push(existing[key].summary + ' (' + key + ')');
  }
  return plan;
}

/** Dry run: logs planned actions, changes nothing. */
function dryRun() {
  const cfg = getConfig();
  const plan = planChanges(cfg);
  Logger.log('=== DRY RUN (%s) ===', new Date());
  Logger.log('Shifts fetched from Jira: %s', plan.creates.length + plan.updates.length);
  Logger.log('To create (%s): %s', plan.creates.length, plan.creates.join('\n  '));
  Logger.log('To update (%s): %s', plan.updates.length, plan.updates.join('\n  '));
  Logger.log('To delete (%s): %s', plan.deletes.length, plan.deletes.join('\n  '));
}

/** Main entry point: create/update/delete events to mirror Jira. */
function syncCalendar() {
  const cfg = getConfig();
  const plan = planChanges(cfg);
  const cal = getOrCreateCalendar(cfg);
  const existing = listSyncedEvents(cal.getId());

  const shifts = fetchAllShifts(cfg);
  const seen = {};
  let created = 0, updated = 0, removed = 0;

  for (const s of shifts) {
    seen[s.shiftId] = true;
    const payload = buildEventPayload(s, cfg);
    const ev = existing[s.shiftId];
    const guestsChanged = payload.attendees.length
      && (!ev || (ev.attendees || []).map(a => a.email).join(',') !== payload.attendees[0].email);
    if (!ev) {
      Calendar.Events.insert(payload, cal.getId(), { sendUpdates: cfg.addGuests && payload.attendees.length ? 'all' : 'none' });
      created++;
      Logger.log('Created: %s', payload.summary);
    } else if (!sameEvent(ev, payload)) {
      Calendar.Events.patch(payload, cal.getId(), ev.id, { sendUpdates: guestsChanged ? 'all' : 'none' });
      updated++;
      Logger.log('Updated: %s', payload.summary);
    }
  }

  for (const key in existing) {
    if (!seen[key]) {
      Calendar.Events.remove(cal.getId(), existing[key].id);
      removed++;
      Logger.log('Removed stale event: %s', existing[key].summary);
    }
  }

  const props = PropertiesService.getScriptProperties();
  props.setProperty('LAST_SYNC', new Date().toISOString());
  Logger.log('Sync done. created=%s updated=%s removed=%s unchanged=%s',
    created, updated, removed, plan.creates.length + plan.updates.length - created - updated);
  return { created, updated, removed };
}

/** Deletes every event this script manages on the calendar (full reset). */
function deleteAllSyncedEvents() {
  const cfg = getConfig();
  const cal = getOrCreateCalendar(cfg);
  const existing = listSyncedEvents(cal.getId());
  let n = 0;
  for (const key in existing) {
    Calendar.Events.remove(cal.getId(), existing[key].id);
    n++;
  }
  Logger.log('Deleted %s synced events from "%s"', n, cfg.calendarName);
}

/* ------------------------------------------------------------------ */
/* Trigger                                                             */
/* ------------------------------------------------------------------ */

/** Installs (or refreshes) the periodic resync trigger. */
function installTrigger() {
  const cfg = getConfig();
  ScriptApp.getProjectTriggers()
    .filter(t => t.getHandlerFunction() === 'syncCalendar')
    .forEach(t => ScriptApp.deleteTrigger(t));
  ScriptApp.newTrigger('syncCalendar')
    .timeBased()
    .everyHours(cfg.triggerHours)
    .create();
  Logger.log('Trigger installed: syncCalendar every %s hours.', cfg.triggerHours);
}

function removeTrigger() {
  ScriptApp.getProjectTriggers()
    .filter(t => t.getHandlerFunction() === 'syncCalendar')
    .forEach(t => ScriptApp.deleteTrigger(t));
  Logger.log('Trigger removed.');
}

/** Convenience: config -> dry run -> sync -> trigger. */
function runAll() {
  setupConfig({});
  dryRun();
  syncCalendar();
  installTrigger();
}
