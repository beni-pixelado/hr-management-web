# Roadmap

> **Reality check:** solo developer, zero budget. Every item below must be buildable on evenings/weekends, shippable in small chunks, and run **100% on free tiers**. If a feature needs a paid service, a second person, or a rewrite, it's cut — or parked in the "later / never" section at the bottom.

This is a time-boxed plan (phases), not version numbers. Versions imply a big-bang release; phases are weekly slices of work you can actually finish alone.

---

## ✅ Already shipped

Don't build any of this again:

- [x] Session auth (gorilla/sessions + bcrypt) + password recovery e-mail
- [x] Candidate CRUD, multi-field search (ILIKE), pagination, status pipeline, ID card
- [x] Departments: create, list, detail, delete, add/remove members
- [x] Overview analytics (pie + line charts via JSON APIs)
- [x] Reports (create + list)
- [x] Account management (profile, Cloudinary photo, password change, account delete)
- [x] Landing page + dark modern design system
- [x] PostgreSQL on Neon free tier, photos on Cloudinary free tier

---

## The one thing that matters right now: roles

Hard truth: today there is only **one role — "logged in"**. There is no story to tell a client, no multi-tenant demo, no reason for anyone to pay. Roles are the difference between "an app I built" and "a product I can hand to someone."

---

## Phase 1 — Make it safe & sellable (weeks 1–4, $0)

Small, boring, high-trust. Do it first.

- [ ] **RBAC: Admin / Recruiter / Viewer** (2–3 weeks) — user table gets a `role` field, middleware checks permissions, UI hides admin actions from viewers. This is the unlock for everything else.
- [ ] **CSRF protection on all POST forms** (few days) — signed token per session; today any site can POST to your forms.
- [ ] **Rate limiting on `/login`, `/register`, `/recuperateaccount`** (few days) — simple in-memory per-IP limiter with a mutex. No Redis needed at this scale (it resets on restart; fine).
- [ ] **Remove the `/debug/cookie` route** (minutes) — it's a dev leftover and a red flag if someone audits the app.

**Exit criteria:** three roles working, no debug routes, brute-force login is slow, all forms validate tokens.

---

## Phase 2 — Core recruiting value (weeks 5–8, $0)

Features a real HR person would miss immediately.

- [ ] **Candidate notes** — free-form text per candidate, with author and timestamp. The single most requested recruiting feature.
- [ ] **Interview scheduling** — a date + status per candidate and an "upcoming interviews" list on the dashboard. One field and one list, not a calendar app.
- [ ] **Attach the interview to the candidate** so the detail page shows the whole pipeline story.

**Explicitly not doing:** iCal export, calendar sync, email reminders. Niche, and Gmail SMTP already covers reminders if a user asks.

---

## Phase 3 — Dev velocity & reliability (weeks 9–12, $0)

You can't afford a QA person, so the codebase has to babysit you.

- [ ] **Real Go tests for core flows** — auth, candidate create/status, department members. `make test` must pass.
- [ ] **GitHub Actions CI** (free) — test → build on every push, deploy to Render on `main`. Catches your mistakes while you sleep.
- [ ] **Docker + Compose for one-command local dev** (free) — app + Postgres up with `docker compose up`. Removes "works on my machine" friction.
- [ ] **Structured logging with stdlib `log/slog`** — zero new dependencies, readable logs on Render.

**Defer until it hurts:** GIN trigram index for search, connection pooling tuning, object storage. Neon + single instance is correct at this scale.

---

## Phase 4 — Launch & find users (weeks 13–16, $0)

A solo dev's only free growth lever is distribution. No money = your time on the keyboard.

- [ ] **Polish onboarding** — empty states, one-click sample-data seeder, a public demo account so anyone can try it without registering.
- [ ] **Live demo link in the README** — Render free tier is fine; cold starts are acceptable for a demo.
- [ ] **Lightweight free analytics** — GoatCounter (free tier) or plain server logs. Know if anyone is even visiting.
- [ ] **Post it everywhere free** — Product Hunt, Hacker News, r/SaaS, r/sideproject, dev.to, LinkedIn. One good writeup of "how I built a $0/mo SaaS" beats a month of features.
- [ ] **Ask the first 10 real users what breaks** — then go back to Phase 1 priorities. The roadmap ends here until real people reply.

---

## Later — only if real users ask for it

Not on the critical path. Park these; each one is a rabbit hole for a solo dev.

- [ ] Audit trail (who changed what/when) — most clients eventually ask; build it then.
- [ ] Full REST API + OpenAPI docs — only if you build a mobile app or third-party integrations.
- [ ] HTMX partial reloads — only if users complain about full-page reloads. The server-rendered app is fine as-is.
- [ ] iCal export / calendar sync — niche; skip unless someone pays for it.

---

## Never (solo + $0) — the honest list

These are traps for a broke solo dev. Do not schedule them.

- **Prometheus / Grafana / APM dashboards** — a free-tier app doesn't need infrastructure dashboards. `slog` + Render logs is enough.
- **Kubernetes / microservices / multi-node anything** — you have no scale problem. Single instance + Neon is correct until it isn't.
- **Redis, Kafka, message queues** — in-memory rate limiting + cookie sessions cover your load.
- **Paid e-mail / paid monitoring / paid AI** — Gmail SMTP (app password, $0), Render logs ($0), no LLM features. Revisit only when revenue exists.

---

## The cost sheet (everything $0/month)

| Service | Job | Free tier reality |
|---|---|---|
| Neon | PostgreSQL | Scales to zero when idle; limits connection count — design for it |
| Render | Hosting | Spins down after ~15 min idle; cold start on first hit |
| Cloudinary | Photo storage | Free quota is enough for this scale |
| Gmail SMTP | Recovery e-mails | ~500/day cap; fine alone |
| GitHub + Actions | Repo + CI | Free private repo + free build minutes |
| **Total** | | **$0.00/month** |

---

## Rules for solo + $0

1. **Free-tier limits are the spec.** Render cold start, Neon connections, Cloudinary quota, Gmail daily cap — design around all four.
2. **Ship small.** At least one deployable change per week. A chunk you can't deploy in a day is too big.
3. **No feature gets built until it earns a user.** If it doesn't help get or keep someone, it's a draft on a branch, not a roadmap item.
4. **Rewrites are the enemy.** Server-rendered Go + vanilla JS + Postgres is already the cheapest stack to run. Don't swap it for the "nicer" version.
5. **When real users show up, this roadmap ends.** Their feedback becomes the roadmap.
