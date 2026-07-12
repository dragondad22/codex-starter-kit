# Decision Records

This directory is the normal reading surface for approved product and architecture
decisions. Start with [INDEX.md](INDEX.md).

The discovery record preserves discussion and source history. A decision record distills
one approved decision into durable context, consequences, and authority. If a record and
its source D-item conflict, stop and reconcile them; do not silently choose one.

Agents normally load the index and only the relevant decision records. The long discovery
record is provenance: follow a record's source breadcrumb only when reviewing origin,
resolving ambiguity, or proposing supersession.

## Required format

Every record contains:

- stable `DEC-NNNN` identity and title;
- `Accepted`, `Superseded`, or `Retired` status;
- owner and decision date;
- exactly one source D-item for the initial D1–D12 promotion;
- context, decision, consequences, and source sections;
- replacement links when superseded.

IDs are never reused. Amendments preserve history. Create a decision only for a choice
that is costly to reverse, surprising without context, and the result of a real tradeoff.
