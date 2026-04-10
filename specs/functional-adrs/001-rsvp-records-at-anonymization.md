# ADR 001 — RSVP Records at Anonymization

| Field | Value |
|-------|-------|
| Status | Accepted |
| Date | 2026-04-11 |

## Context

When an admin anonymizes a musician account, the vision explicitly defines what happens to
FeePayment records (anonymized: account name replaced with an opaque token; financial data
retained). It is silent on RSVP records.

Two interpretations are equally simple:
1. **Delete** RSVP records at anonymization time.
2. **Anonymize** RSVP records (replace account name with a token or generic label), analogous to
   fee payments.

The choice affects the information model, the anonymization procedure, and what data remains
after anonymization.

## Decision

**Delete all RSVP records belonging to the account at anonymization time.**

## Rationale

1. **The processing purpose is identity-bound.** RSVP records exist for event management — to
   know who plans to attend so the association can organize accordingly. Once the musician's
   identity is erased, the organizational value of their RSVP records is gone. A count of
   anonymous yes/no/maybe responses has no operational use in V1 (the statistics feature is
   deferred).

2. **Unlike FeePayments, there is no retention obligation.** Fee payment records are kept
   because they carry financial history (amount, season, type) that remains meaningful in
   aggregate even without identity. RSVP records carry no equivalent aggregate value.

3. **Consistency with the event-deletion model.** The vision establishes event deletion as the
   primary GDPR compliance path for RSVP records. Treating account anonymization the same way
   (delete the RSVPs) is the simplest extension of this model.

4. **Art. 17 overrides Art. 6(1)(f).** The 2-year retention period for RSVP records is based on
   legitimate interest. When a data subject requests erasure (implemented here as
   anonymization), the legitimate interest in retaining their RSVP records does not outweigh
   their right to erasure — the purpose the legitimate interest serves (event management) has
   no continuing value without identity.

## Consequences

- The anonymization procedure in the Musician Management spec deletes all RSVP records for the
  account.
- After anonymization, affected past events show fewer entries in their RSVP lists, with no trace
  of the anonymized account.
- If a statistics feature is built in V2 that needs historical attendance data, this decision
  will need to be revisited. At that point, the retention basis and the value of anonymized RSVP
  records should be re-evaluated.

## Alternatives Considered

**Keep RSVP records with an anonymization token:** Rejected. The event management purpose has no
value without identity. Unlike fee payments, there is no financial-record rationale for keeping
the data. Adding a token to RSVP records increases complexity without benefit in V1.
