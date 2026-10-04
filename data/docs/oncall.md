# On-call Handbook

Engineers on the platform team share a weekly on-call rotation for the Kestrel API.

## Rotation

Each on-call shift lasts one week. The handoff happens every Monday at 10:00 UTC in the #oncall-handoff channel, where the outgoing engineer summarises open incidents.

## Response times

A page must be acknowledged within 15 minutes, at any hour. If the primary engineer does not acknowledge, the page escalates to the secondary engineer after 15 minutes, and to the engineering manager after 30 minutes.

## Severity levels

- SEV1: the Kestrel API is down or losing data for all customers. Start an incident call immediately and post updates every 30 minutes.
- SEV2: a major feature is broken or a large customer is affected. Post updates every 2 hours.
- SEV3: a minor bug or degraded performance with a workaround. Handle during working hours.

## Compensation

On-call engineers receive a stipend of 300 USD per week of on-call. Any time spent handling a page between 22:00 and 07:00 local time can be taken back as time off within the following month.

## Postmortems

Every SEV1 and SEV2 incident needs a written postmortem within 5 working days. Postmortems are blameless and focus on systems, not people.
