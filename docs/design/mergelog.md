## Merge Log

### Definition

- **Safe**: A log that is safe if it has been replicated to a quorum, no matter
  whether or not the **committed** flag is set on any replica.

---

After a replica established leadership, it needs to merge the latest logs from a
quorum of replicas, to ensure that the leader replica has all **safe** logs.


### Lemma-only-max-committer-logs

In TRaft, a leader only needs to use the logs from replicas with the max **committer**(`(term, id)`).

A log that does not present in any max-committer replicas a leader seen can **not**
be safe.

Proof:

If a log `A` becomes **safe**, it will be seen by the next leader `Li`.
Because a leader has to collect logs from a quorum and any two quorum
intersections with each other.

If the next next leader `Lj`(`j>i`) has seen `Li`, it will choose `A`.
Otherwise, it will see a replica that has `A`.

∴ Any newer leader will choose `A`.

∴ TRaft only need to merge logs from replicas with the latest `Committer`.

E.g.:

```
Li: indicates a replica becomes leader
A/B: a log is written

R0 L1 A
R1    A
R2    A
R3      L2 B
R4           L3
------------------------> time
```

In this digram:
L2 sees `A` and then updates its `Committer` to `L2`, then writes log `B` to its
local log. And R3 crashed before forwarding any log out.

-   If L3 established its leadership by contacting R2 and R3, it uses logs from only
    R3(R3 has the latest committer `L2`), it will see `A`.

-   If L3 established its leadership by contacting R0 and R1, it will also see `A`.

