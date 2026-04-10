# Future Compatibility

Assessment of each Later feature from `specs/goals/vision.md` against the V1 architecture.
For each feature: current fit, the main architectural consideration, and where the biggest
change would land.

---

## Public Face — Blog and Articles

**Current fit:** Partial. Routing and schema are ready; the Go/Buffalo stack is workable
but not CMS-native.

**Architectural fit:**
- Routing is fully extensible. Public routes follow a clear pattern (`GET /`, `GET /about`,
  etc.); blog routes (`GET /articles`, `GET /articles/{slug}`) add a new public namespace
  without touching anything existing.
- Schema is purely additive: an `articles` table (slug, title, body, published_at,
  author_id) references `accounts` for authorship. No existing table changes.
- ADR 005's roles table is ready for a `content_editor` role with zero schema work.
  Application-layer work (middleware, route group, UI) would be required — see the ADR 005
  architectural issue.

**Friction points (Go/Buffalo is not CMS-native):**
- No built-in file upload abstraction. Images in articles require manual integration with
  OVH Object Storage (presigned URLs, upload handler). This is 1–3 days of infrastructure
  work that a CMS platform would provide out of the box.
- No generated admin panel. The article editing UI is hand-written Plush templates. A
  WYSIWYG experience requires integrating a JavaScript rich-text editor (e.g. Trix,
  Quill) as a static asset.
- Slug generation and uniqueness enforcement are application-level responsibilities.

**Scope risk:** Simple articles (create/edit/publish) are straightforward. If the scope
expands to media libraries, scheduled publishing, or multi-author workflows, the stack
becomes a sustained friction point compared to a dedicated CMS. The decision of whether
to keep this in Buffalo or extract to a separate tool should be revisited when the
Community Manager role is defined.

**Biggest change:** The JS rich-text editor and file upload pipeline. Everything else is
additive.

---

## Statistics Page

**Current fit:** Strong. No structural work required.

**Architectural fit:**
- All data needed for a statistics page is already captured correctly: `fee_payments`
  links accounts to seasons (with amount and type), `accounts` stores `birth_date` and
  first inscription date is derivable, `rsvps` record attendance per event with instrument.
- A statistics page is a new authenticated route (`GET /statistiques`) returning aggregated
  queries. No schema changes, no new tables, no middleware changes.
- The most complex query (age pyramid per season) is a straightforward GROUP BY on
  `accounts.birth_date` filtered by `fee_payments.season_id`. All required columns exist.

**Biggest change:** Writing the aggregation queries and the display template. Pure
application-layer addition.

---

## Improved Account Model

**Current fit:** Blocked by a V1 technical requirement.

**Architectural fit:**
- V1 deliberately has no system email infrastructure. Invite and password-reset links are
  manually copied and sent by an admin. Any meaningful improvement to the account model
  (self-service registration, automated invites, email-based password reset, email
  verification) requires reversing this requirement.
- Reversing it touches: environment variables (SMTP credentials), configuration spec,
  deployment (outbound email from the VPS or a relay service), and GDPR (email as a
  communication channel with new data-flow implications).
- The rest of the account model is clean and extensible. Additional profile fields are
  additive. State machine transitions (pending → active → anonymized) can be extended
  without breaking existing states.

**Biggest change:** Introducing email infrastructure. This is a deployment and
configuration decision first, not a schema or routing decision.

---

## Sheet Music Search and Database

**Current fit:** Additive. Current implementation (Google Drive link) leaves no technical
debt that blocks this.

**Architectural fit:**
- V1 stores the sheet music link as a single environment variable. An integrated
  experience requires a new schema (pieces table: title, composer, instrumentation,
  file reference) and file storage.
- Storage strategy is the open decision: VPS local volume (simple, no external dependency,
  limited scalability) versus OVH Object Storage (already used for backups, consistent
  infrastructure, requires presigned URL handling). This decision does not need to be made
  now.
- Search is a standard Go pattern: a `ILIKE`-based query on title/composer for simple
  needs, or a PostgreSQL full-text index (`tsvector`) if performance requires it.
  Either can be added without touching existing tables.

**Biggest change:** The storage strategy decision and the file upload pipeline (same
infrastructure concern as the Blog feature). If both features are implemented together,
the upload infrastructure is built once.

---

## Trombinoscope

**Current fit:** Additive schema, but requires a new GDPR consent field.

**Architectural fit:**
- A photo for each musician is a single additive column on `accounts`
  (`photo_url VARCHAR` or a foreign key to a photos table). No existing columns change.
- GDPR is the real constraint. Photo display constitutes processing of biometric data
  under GDPR. A trombinoscope requires a separate, explicit consent distinct from the
  existing phone/address consent. The consent collection point is the invite flow — adding
  a third consent checkbox requires a spec update to the invite form and a new boolean
  column on `accounts`.
- Consent withdrawal must clear the photo, following the same pattern as phone/address.
  This extends the anonymization transaction by one step.
- File upload for photos is the same infrastructure concern as Blog and Sheet Music.

**Biggest change:** The GDPR consent model extension and the upload pipeline. The
display is trivial once those are in place.

---

## Commission Artistique

**Current fit:** Purely additive. Current specs have no conflict with this feature.

**Architectural fit:**
- Scope is undefined in V1. Possible interpretations range from a shared document space
  (documents table linking to OVH Object Storage) to a lightweight discussion board
  (threads + messages tables).
- ADR 005's roles table is the natural home for a `commission_artistique` role, limiting
  access to designated members. Zero schema work at the role layer.
- Any implementation is purely additive to the current schema and routing structure.

**Biggest change:** Scope definition. Architectural impact depends entirely on what
"tooling around it" means in practice.

---

## Assemblées Générales and Legal Documents

**Current fit:** Purely additive. No conflict with current specs.

**Architectural fit:**
- The simplest implementation is a documents table (title, file_url, document_type,
  published_at) with an authenticated route listing documents by type. Files live in
  OVH Object Storage (consistent with the backup infrastructure already in place).
- Access control is the only question: all authenticated musicians, or admins only?
  Either is a one-line middleware change on the route group.
- No GDPR implications (legal documents are institutional, not personal data).

**Biggest change:** The file upload pipeline — same infrastructure concern shared with
Blog, Sheet Music, and Trombinoscope. These four features should be planned together
to avoid implementing the same pipeline four times.

---

## Cross-Cutting Note: File Upload Infrastructure

Four Later features (Blog, Sheet Music, Trombinoscope, Assemblées Générales) all require
file upload to OVH Object Storage. None of this infrastructure exists in V1. When any of
these features is picked up, the upload pipeline should be designed as shared
infrastructure rather than built four times in isolation.

The V1 backup script already uses OVH Object Storage with AWS-compatible credentials.
The same bucket configuration (or a second bucket) is the natural starting point.
