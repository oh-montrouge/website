# Adj-4 — Test Coverage Gaps

**Status:** Draft

**Goal:** Close the unit/integration test gaps that pull actions (63%), models (35%), and
services (52%) below the expected near-100% coverage.

---

## Exempt Files

`actions/app.go` (0%) — legitimately exempt per CLAUDE.md: "Files whose sole content is
an init() for framework wiring (e.g. render.go, app.go)". The `App()` function is pure
Buffalo router/middleware wiring; `translations()` and `forceSSL()` are thin framework
wrappers with no domain logic. No tests needed.

`services/repositories.go` (not in coverage) — interface-only file, no method bodies.
Correctly excluded.

---

## Gap 1 — Missing test files (CLAUDE.md violation) ✅ Done

Three model files have no `*_test.go` at all. Each contains concrete method bodies on a
`Store` type, so none qualifies for any exemption.

### `models/event_test.go` — does not exist

**EventStore** (integration tests against real DB, follow pattern of `fee_payment_test.go`):

| Test | What to assert |
|------|----------------|
| `TestEventStore_Create` | Returns a new ID; row readable via GetByID |
| `TestEventStore_GetByID_Found` | Returns correct row |
| `TestEventStore_GetByID_NotFound` | Returns `nil, nil` |
| `TestEventStore_Update` | Name/type/datetime change persisted |
| `TestEventStore_Delete` | Row gone; subsequent GetByID returns nil |
| `TestEventStore_ListUpcoming_Empty` | Returns empty slice when no events |
| `TestEventStore_ListUpcoming_FiltersOld` | Past events excluded; upcoming included |
| `TestEventStore_ListUpcoming_IncludesRSVPState` | RSVP state correct for viewer; null for non-viewer |
| `TestEventStore_ListAll_IncludesPastAndFuture` | Both past and future events returned |
| `TestEventStore_DeleteFields` | Cascades to event_fields rows |
| `TestEventStore_AddField_GetFieldByID` | Round-trip: add then fetch |
| `TestEventStore_GetFieldByID_NotFound` | Returns `nil, nil` |
| `TestEventStore_UpdateField` | Label/type/required/position persisted |
| `TestEventStore_DeleteField` | Field gone after delete |
| `TestEventStore_ListFields` | Ordered by position |
| `TestEventStore_ListFieldChoices` | Ordered by position |
| `TestEventStore_CountFieldResponses_Zero` | Returns 0 when none |
| `TestEventStore_CountFieldResponses_NonZero` | Returns correct count after insert |
| `TestEventStore_AddFieldChoice` | Choice readable after add |
| `TestEventStore_DeleteFieldChoices` | All choices removed |

**RSVPStore** (integration tests):

| Test | What to assert |
|------|----------------|
| `TestRSVPStore_SeedForEvent` | Creates one RSVP per active account; idempotent (ON CONFLICT) |
| `TestRSVPStore_SeedForAccount` | Creates one RSVP per future event; idempotent |
| `TestRSVPStore_GetByAccountAndEvent_Found` | Returns correct RSVP |
| `TestRSVPStore_GetByAccountAndEvent_NotFound` | Returns `nil, nil` |
| `TestRSVPStore_Update_NilInstrument` | Sets instrument_id to NULL |
| `TestRSVPStore_Update_WithInstrument` | Sets instrument_id to provided value |
| `TestRSVPStore_DeleteByAccount` | All RSVPs for account removed |
| `TestRSVPStore_ClearFieldResponses` | Responses for RSVP removed; other RSVP untouched |
| `TestRSVPStore_ListForEvent` | Returns all RSVPs with joined account/instrument fields |
| `TestRSVPStore_ResetYesRSVPs` | Yes RSVPs → unanswered with null instrument; no/maybe untouched |
| `TestRSVPStore_ClearInstruments` | instrument_id set to NULL where not already null |
| `TestRSVPStore_AddFieldResponse_InsertAndUpsert` | Insert works; second call with same rsvp+field updates value |
| `TestRSVPStore_ListFieldResponses` | Returns responses with field label/type; ordered by position |

---

### `models/role_test.go` — does not exist

**AccountRoleStore** (integration tests):

| Test | What to assert |
|------|----------------|
| `TestAccountRoleStore_HasRole_True` | Returns true when role assigned |
| `TestAccountRoleStore_HasRole_False` | Returns false when role not assigned |
| `TestAccountRoleStore_HasActiveRoleHolder_True` | True when active admin exists |
| `TestAccountRoleStore_HasActiveRoleHolder_False` | False when no active admin |
| `TestAccountRoleStore_GetIDByName` | Returns correct role ID |
| `TestAccountRoleStore_AssignRole` | HasRole true after assign |
| `TestAccountRoleStore_CountActiveAdmins` | Returns 0 initially; 1 after assign |
| `TestAccountRoleStore_RevokeRole` | HasRole false after revoke |
| `TestAccountRoleStore_RemoveAllRoles` | All roles removed; HasRole false for each |

---

### `models/session_test.go` — does not exist

**HTTPSessionStore** (integration tests):

| Test | What to assert |
|------|----------------|
| `TestHTTPSessionStore_BindAccount` | account_id set on session row |
| `TestHTTPSessionStore_DeleteByAccount` | Session rows for account gone |

Note: tests require an `http_sessions` row to exist first — check test harness for how
to insert raw rows if there is no Store factory for sessions.

---

## Gap 2 — `models/account.go` (30%)

**File:** `webapp/models/account_test.go`
**Currently tested:** `FindByEmail` (found + not-found), `GetByID` (found + not-found).
**Missing — 14 untested methods:**

| Test | Key assertions |
|------|----------------|
| `TestAccountStore_Create` | Returns ID; row present with correct status='active' |
| `TestAccountStore_UpdatePasswordHash` | Hash column updated |
| `TestAccountStore_Activate_WithConsent` | Status='active'; phone/address preserved |
| `TestAccountStore_Activate_WithoutConsent` | Status='active'; phone and address set to NULL |
| `TestAccountStore_CreatePending` | status='pending'; phone_address_consent=false; processing_restricted=false |
| `TestAccountStore_UpdateEmail` | Email column updated |
| `TestAccountStore_Delete` | Row gone; subsequent GetByID errors |
| `TestAccountStore_AnonymizeAccount` | email=NULL; password_hash=NULL; status='anonymized'; token set |
| `TestAccountStore_GetProfile` | Returns joined instrument name; all nullable fields |
| `TestAccountStore_SetProfile` | first_name, last_name, birth_date, parental_consent_uri persisted |
| `TestAccountStore_UpdateProfile` | All 8 columns updated in single call |
| `TestAccountStore_ListNonAnonymized` | Includes active accounts; excludes anonymized; is_admin correct |
| `TestAccountStore_ListForRetentionReview` | Only returns accounts whose last season ended >5 years ago |
| `TestAccountStore_ClearMembershipFields` | All personal fields set to NULL; booleans reset to false |
| `TestAccountStore_WithdrawConsent` | phone=NULL; address=NULL; phone_address_consent=false |
| `TestAccountStore_ToggleProcessingRestriction` | false→true on first call; true→false on second |

---

## Gap 3 — `services/event.go` (30%)

**File:** `webapp/services/event_test.go`
**Currently tested:** Create, Update (all 5 type-change cases), UpdateRSVP (6 cases),
UpdateField (blocked + allowed), DeleteField (blocked + allowed), AddField (other + non-other),
SeedRSVPsForAccount.
**Missing — 10 untested functions:**

### Pure helpers (no stubs required, table-driven)

| Test | What to assert |
|------|----------------|
| `TestOwnRSVPDTO_FieldValue_Found` | Returns Value for matching FieldID |
| `TestOwnRSVPDTO_FieldValue_NotFound` | Returns `""` when fieldID absent |
| `TestRSVPRowDTO_FieldValueMap_Empty` | Returns empty map |
| `TestRSVPRowDTO_FieldValueMap_Populated` | Each fieldID (as string key) maps to Value |
| `TestToSummaryDTOs_NullRSVPState` | RSVPState mapped to `""` when nulls.String not Valid |
| `TestToSummaryDTOs_ValidRSVPState` | RSVPState propagated from String field |
| `TestDisplayName_FullName` | Returns "LastName FirstName" when both valid |
| `TestDisplayName_Anonymized` | Returns "Musicien " + first 8 chars of token when names null |
| `TestDisplayName_Unknown` | Returns "Compte inconnu" when neither valid |
| `TestBuildPupitre_UsesRSVPInstrumentForYes` | Yes RSVPs counted under RSVPInstrumentName |
| `TestBuildPupitre_UsesMainInstrumentForNonYes` | No/maybe/unanswered counted under MainInstrumentName |
| `TestBuildPupitre_AllStates` | Yes/Maybe/No/Unanswered counts correct |
| `TestBuildPupitre_PreservesOrder` | Row order matches first-seen instrument order |
| `TestBuildPupitre_Empty` | Returns empty slice |

### Read operations (use existing stubs in event_test.go)

| Test | Scenario | Key assertions |
|------|----------|----------------|
| `TestEventService_ListForMember_Success` | Repo returns rows | DTOs match; RSVPState propagated |
| `TestEventService_ListForMember_RepoError` | Repo returns error | Error propagated |
| `TestEventService_ListAll_Success` | Repo returns rows | DTOs returned |
| `TestEventService_ListAll_RepoError` | Repo returns error | Error propagated |
| `TestEventService_GetDetail_EventNotFound` | GetByID returns nil | `ErrEventNotFound` |
| `TestEventService_GetDetail_GetByIDError` | GetByID returns error | Error propagated |
| `TestEventService_GetDetail_RSVPsError` | ListForEvent returns error | Error propagated |
| `TestEventService_GetDetail_OwnRSVPError` | GetByAccountAndEvent returns error | Error propagated |
| `TestEventService_GetDetail_OtherEvent_LoadsFields` | event_type=other; loadFields called | Fields in DTO |
| `TestEventService_GetDetail_OtherEvent_LoadsFieldResponses` | event_type=other | FieldResponses in RSVPs |
| `TestEventService_GetDetail_ConcertEvent_NoFields` | event_type=concert | Fields slice empty |
| `TestEventService_GetDetail_WithOwnRSVP` | GetByAccountAndEvent returns RSVP | OwnRSVP populated with instrument |
| `TestEventService_GetDetail_WithoutOwnRSVP` | GetByAccountAndEvent returns nil | OwnRSVP is nil |
| `TestEventService_GetDetail_NullableDescription` | Description not valid | DTO.Description is `""` |
| `TestEventService_GetField_NotFound` | GetFieldByID returns nil | `ErrEventFieldNotFound` |
| `TestEventService_GetField_NonChoiceType` | field_type=text | Choices slice empty |
| `TestEventService_GetField_ChoiceType` | field_type=choice | ListFieldChoices called; Choices populated |
| `TestEventService_GetField_ChoiceTypeError` | ListFieldChoices returns error | Error propagated |

---

## Secondary gaps (not requested, recorded for completeness)

These files have partial coverage but were not part of the focused analysis above.
Review separately if coverage targets are set.

| File | Coverage | Likely cause |
|------|----------|--------------|
| `actions/events.go` | 43% | Error paths and edge cases in RSVP/field handlers untested |
| `services/account.go` | 63% | Several branches in CompleteInvite / UpdatePassword untested |
| `actions/musicians.go` | 64% | Admin-only handlers not fully exercised |
| `actions/tokens.go` | 75% | Reset/invite expiry/not-found paths missing |
| `services/musician.go` | 68% | Some compliance/anonymization branches |

---

## Implementation notes

- All `models/*_test.go` integration tests use the shared `DB` handle and `truncateAll(t)`
  fixture — follow the pattern in `models/fee_payment_test.go`.
- Service tests use stub repositories already defined in `services/event_test.go` and
  `services/account_test.go` — extend or clone as needed.
- Pure-function tests (Gap 3 helper section) need no DB and no stubs; single file,
  table-driven, fast.
- `models/session_test.go` may need direct SQL inserts into `http_sessions` if no Store
  factory exists for that table — check `truncateAll` scope first.

## Definition of Done

At the end of each implementation work, the pre-commit lefthook must pass.
