# Adj-4 — Test Coverage Gaps

**Status:** Complete

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

## Gap 2 — `models/account.go` (30%) ✅ Done

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

## Gap 3 — `services/event.go` (30%) ✅ Done

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

---

## Gap 4 — `actions/events.go` (43%) ✅ Done

**File:** `webapp/actions/events_test.go`
**Currently tested:** Dashboard (3 cases), Index (1 case, no error path), Show (3 cases),
UpdateRSVP (4 cases), New, Create (4 cases), Delete (2 cases), AdminUpdateRSVP (5 cases).
**Missing — 6 handlers entirely untested + minor gaps on existing handlers:**

### Untested handlers

#### `Edit`

| Test | Key assertions |
|------|----------------|
| `TestEventsHandler_Edit_RendersForm` | 200; event name in body |
| `TestEventsHandler_Edit_InvalidID` | non-numeric `:id` → 404 |
| `TestEventsHandler_Edit_EventNotFound` | `detailErr = ErrEventNotFound` → 404 |

#### `Update`

| Test | Key assertions |
|------|----------------|
| `TestEventsHandler_Update_Success` | 303 → `/admin/evenements` |
| `TestEventsHandler_Update_InvalidID` | non-numeric `:id` → 404 |
| `TestEventsHandler_Update_MissingName` | empty name → 422; event still set on context |
| `TestEventsHandler_Update_InvalidDate` | bad date → 422; event still set on context |
| `TestEventsHandler_Update_EventNotFound` | `updateErr = ErrEventNotFound` → 404 |

Note: The 422 branches call `h.Events.GetDetail` to re-render the form; stub must return
a non-nil `detail` for those cases.

#### `AddField`

| Test | Key assertions |
|------|----------------|
| `TestEventsHandler_AddField_Success` | 303 → `/admin/evenements/1/modifier` |
| `TestEventsHandler_AddField_InvalidID` | non-numeric `:id` → 400 |
| `TestEventsHandler_AddField_FieldOnlyForOther` | `addFieldErr = ErrFieldOnlyForOther` → 303 (flash) |
| `TestEventsHandler_AddField_ServiceError` | other error → 500 |

#### `EditFieldForm`

| Test | Key assertions |
|------|----------------|
| `TestEventsHandler_EditFieldForm_RendersForm` | 200; field label in body |
| `TestEventsHandler_EditFieldForm_InvalidFieldID` | non-numeric `field_id` → 404 |
| `TestEventsHandler_EditFieldForm_InvalidEventID` | non-numeric `:id` → 404 |
| `TestEventsHandler_EditFieldForm_FieldNotFound` | `fieldErr = ErrEventFieldNotFound` → 404 |

#### `UpdateField`

| Test | Key assertions |
|------|----------------|
| `TestEventsHandler_UpdateField_Success` | 303; flash success |
| `TestEventsHandler_UpdateField_FieldHasResponses` | `updateFieldErr = ErrFieldHasResponses` → 303 (flash) |
| `TestEventsHandler_UpdateField_ServiceError` | other error → 500 |

#### `DeleteField`

| Test | Key assertions |
|------|----------------|
| `TestEventsHandler_DeleteField_Success` | 303; flash success |
| `TestEventsHandler_DeleteField_FieldHasResponses` | `deleteFieldErr = ErrFieldHasResponses` → 303 (flash) |
| `TestEventsHandler_DeleteField_ServiceError` | other error → 500 |

### Gaps on already-tested handlers

| Test | Key assertions |
|------|----------------|
| `TestEventsHandler_Index_ServiceError` | `listErr = errors.New("db")` → 500 |
| `TestEventsHandler_AdminUpdateRSVP_InvalidMusicianID` | non-numeric `musician_id` → 400 |
| `TestEventsHandler_AdminUpdateRSVP_InvalidBody` | malformed JSON body → 400 |

---

## Gap 5 — `services/account.go` (63%) ✅ Done

**File:** `webapp/services/account_test.go`
**Currently tested:** Authenticate (6), CreateAdmin (2), ResetPassword (3),
GenerateInviteToken (2), ValidateInviteToken (3), CompleteInvite (5),
GeneratePasswordResetToken (1), ValidatePasswordResetToken (2),
CompletePasswordReset (1), ValidatePasswordStrength (6).
**Missing — 8 untested methods + error paths on tested methods:**

**Stub note:** `stubRoleRepo.HasRole` is hardcoded `return false, nil`. Tests for
`GrantAdmin` (already-has-role path) and `RevokeAdmin` (not-admin path) need a
controllable `hasRole bool` field — extend the stub before writing those cases.

### Thin delegating methods

| Test | Key assertions |
|------|----------------|
| `TestGetByID_Success` | returns mapped `AccountDTO` |
| `TestGetByID_Error` | propagates repo error |
| `TestCreatePending_Success` | returns ID from repo |
| `TestCreatePending_Error` | propagates repo error |
| `TestGetActiveInviteToken_NotFound` | repo returns nil → service returns nil |
| `TestGetActiveInviteToken_Found` | URL contains `/invitation/`; ExpiresAt propagated |
| `TestGetActiveInviteToken_Error` | repo error propagated |
| `TestGetActivePasswordResetToken_NotFound` | repo returns nil → service returns nil |
| `TestGetActivePasswordResetToken_Found` | URL contains `/reinitialiser-mot-de-passe/` |
| `TestGetActivePasswordResetToken_Error` | repo error propagated |

### `GrantAdmin`

| Test | Key assertions |
|------|----------------|
| `TestGrantAdmin_AlreadyAdmin_Idempotent` | `HasRole = true` → returns nil, no AssignRole call |
| `TestGrantAdmin_Success` | `HasRole = false` → GetIDByName + AssignRole called |
| `TestGrantAdmin_HasRoleError` | HasRole error propagated |
| `TestGrantAdmin_GetIDByNameError` | GetIDByName error propagated |
| `TestGrantAdmin_AssignError` | AssignRole error propagated |

### `RevokeAdmin`

| Test | Key assertions |
|------|----------------|
| `TestRevokeAdmin_NotAdmin_Idempotent` | `HasRole = false` → returns nil |
| `TestRevokeAdmin_LastAdmin` | count ≤ 1 → `ErrLastAdmin` |
| `TestRevokeAdmin_Success` | count > 1 → RevokeRole called |
| `TestRevokeAdmin_HasRoleError` | propagated |
| `TestRevokeAdmin_CountError` | propagated |
| `TestRevokeAdmin_GetIDByNameError` | propagated |
| `TestRevokeAdmin_RevokeError` | propagated |

### `DeletePending`

| Test | Key assertions |
|------|----------------|
| `TestDeletePending_Success` | status=pending, not last admin → Delete called |
| `TestDeletePending_NotPending` | status=active → `ErrAccountNotPending` |
| `TestDeletePending_LastAdmin` | is admin, count ≤ 1 → `ErrLastAdmin` |
| `TestDeletePending_GetByIDError` | GetByID error propagated |
| `TestDeletePending_DeleteError` | Delete error propagated |

### Error paths on already-tested methods

| Test | Key assertions |
|------|----------------|
| `TestGeneratePasswordResetToken_InvalidateError` | InvalidateExisting error propagated |
| `TestCompletePasswordReset_UpdateHashError` | UpdatePasswordHash error propagated |
| `TestCompletePasswordReset_MarkUsedError` | MarkUsed error propagated |

---

## Gap 6 — `actions/musicians.go` (64%) ✅ Done

**File:** `webapp/actions/musicians_test.go`
**Currently tested:** Index (2), Show (2), Create (3), Delete (2), Anonymize (2),
Edit (2), Update (2), GrantAdmin (1), RevokeAdmin (2), WithdrawConsent (1),
ToggleProcessingRestriction (1), GenerateInviteLink (1), GenerateResetLink (1).
**Missing — `New` handler entirely untested + error/edge-case paths across all handlers:**

### `New` (not tested at all)

| Test | Key assertions |
|------|----------------|
| `TestMusiciansHandler_New_RendersForm` | 200; instrument list in context |
| `TestMusiciansHandler_New_InstrumentsError` | Instruments.List error → 500 |

### `Show`

| Test | Key assertions |
|------|----------------|
| `TestMusiciansHandler_Show_InvalidID` | non-numeric `:id` → 404 |
| `TestMusiciansHandler_Show_ProfileError` | GetProfile error → 500 |
| `TestMusiciansHandler_Show_IsAdminError` | IsAdmin error → 500 |

### `Create`

| Test | Key assertions |
|------|----------------|
| `TestMusiciansHandler_Create_InvalidBirthDate` | unparseable date → 422; "invalide" in body |
| `TestMusiciansHandler_Create_CreatePendingError` | CreatePending error → 422; error message in body |
| `TestMusiciansHandler_Create_GenerateInviteTokenError` | GenerateInviteToken error → 500 |

### `Edit`

| Test | Key assertions |
|------|----------------|
| `TestMusiciansHandler_Edit_InvalidID` | non-numeric `:id` → 404 |

### `Update`

| Test | Key assertions |
|------|----------------|
| `TestMusiciansHandler_Update_InvalidID` | non-numeric `:id` → 404 |
| `TestMusiciansHandler_Update_InvalidBirthDate` | unparseable date → 422 |
| `TestMusiciansHandler_Update_InvalidInstrumentID` | non-numeric instrument_id → 422 |
| `TestMusiciansHandler_Update_ServiceError` | non-parental UpdateProfile error → 422; error in body |

### `Delete`

| Test | Key assertions |
|------|----------------|
| `TestMusiciansHandler_Delete_InvalidID` | non-numeric `:id` → 404 |
| `TestMusiciansHandler_Delete_LastAdmin` | `deleteErr = ErrLastAdmin` → 303 with flash |

### `Anonymize`

| Test | Key assertions |
|------|----------------|
| `TestMusiciansHandler_Anonymize_InvalidID` | non-numeric `:id` → 404 |
| `TestMusiciansHandler_Anonymize_LastAdmin` | `anonymizeErr = ErrLastAdmin` → 303 with flash |

### `GrantAdmin`

| Test | Key assertions |
|------|----------------|
| `TestMusiciansHandler_GrantAdmin_InvalidID` | non-numeric `:id` → 404 |
| `TestMusiciansHandler_GrantAdmin_ServiceError` | grantErr set → 303 with danger flash |

### `RevokeAdmin`

| Test | Key assertions |
|------|----------------|
| `TestMusiciansHandler_RevokeAdmin_InvalidID` | non-numeric `:id` → 404 |
| `TestMusiciansHandler_RevokeAdmin_OtherError` | non-ErrLastAdmin revokeErr → 500 |

### `GenerateInviteLink` / `GenerateResetLink`

| Test | Key assertions |
|------|----------------|
| `TestMusiciansHandler_GenerateInviteLink_InvalidID` | non-numeric `:id` → 404 |
| `TestMusiciansHandler_GenerateInviteLink_ServiceError` | inviteErr set → 500 |
| `TestMusiciansHandler_GenerateResetLink_InvalidID` | non-numeric `:id` → 404 |
| `TestMusiciansHandler_GenerateResetLink_ServiceError` | resetErr set → 500 |

### `WithdrawConsent` / `ToggleProcessingRestriction`

| Test | Key assertions |
|------|----------------|
| `TestMusiciansHandler_WithdrawConsent_InvalidID` | non-numeric `:id` → 404 |
| `TestMusiciansHandler_WithdrawConsent_ServiceError` | consentErr → 500 |
| `TestMusiciansHandler_ToggleProcessingRestriction_InvalidID` | non-numeric `:id` → 404 |
| `TestMusiciansHandler_ToggleProcessingRestriction_ServiceError` | toggleErr → 500 |

---

## Gap 7 — `actions/tokens.go` (75%) ✅ Done

**File:** `webapp/actions/tokens_test.go`
**Currently tested:** InviteForm (3), InviteSubmit (5), ResetForm (2), ResetSubmit (3).
**Missing — 5 paths:**

| Test | Key assertions |
|------|----------------|
| `TestTokensHandler_ResetForm_DBError` | non-ErrInvalidToken validateErr → 500 |
| `TestTokensHandler_ResetSubmit_WeakPassword` | short password → 422; "caractères" in body |
| `TestTokensHandler_ResetSubmit_DBError` | non-ErrInvalidToken validateErr → 500 |
| `TestTokensHandler_ResetSubmit_CompleteError` | `completeErr = errors.New("db")` → 500 |
| `TestTokensHandler_InviteSubmit_CompleteError` | `completeErr = errors.New("db")` → 500 |

---

## Gap 8 — `services/musician.go` (68%) ✅ Done

**File:** `webapp/services/musician_test.go`
**Currently tested:** SetInitialProfile (4 consent-rule cases), UpdateProfile (2 consent-rule
cases), ConsentWithdrawal (2), ListNonAnonymized (1).
**Missing — `GetProfile` and `ToggleProcessingRestriction` entirely untested, plus error paths:**

### `GetProfile`

| Test | Key assertions |
|------|----------------|
| `TestGetProfile_Success_WithBirthDate` | `BirthDate.Valid = true` → `p.BirthDate` non-nil; all string fields mapped |
| `TestGetProfile_Success_NoBirthDate` | `BirthDate.Valid = false` → `p.BirthDate` nil |
| `TestGetProfile_Error` | repo error propagated; nil profile |

### `ToggleProcessingRestriction`

| Test | Key assertions |
|------|----------------|
| `TestToggleProcessingRestriction_Success` | no error |
| `TestToggleProcessingRestriction_Error` | repo error propagated |

### Error paths on already-tested methods

| Test | Key assertions |
|------|----------------|
| `TestListNonAnonymized_Error` | repo error propagated; nil slice |

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
- Gap 5 (`services/account.go`): extend `stubRoleRepo` with `hasRole bool` and
  `hasRoleErr error` fields before writing `GrantAdmin`/`RevokeAdmin` tests —
  `HasRole` is currently hardcoded `return false, nil`.
- Gap 4 (`actions/events.go`): the `Update` 422 branches call `h.Events.GetDetail`
  internally to populate the re-rendered form; ensure `stubEventManager.detail` is
  non-nil in those test cases.

## Definition of Done

At the end of each implementation work, the pre-commit lefthook must pass.
