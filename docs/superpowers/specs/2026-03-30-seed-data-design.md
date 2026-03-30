# Xpressgo Seed Data Design

## Purpose

Define a richer phase-1 demo dataset so the admin panel feels operational and the mini app discovery flow feels populated.

This design updates seed expectations only. It does not change the data model or permission model.

## Goals

- make the admin dashboard, orders, branches, menu, and staff screens feel live
- make discovery feel like a real multi-store platform
- preserve the current branch-aware architecture
- keep the seed command idempotent so rerunning it does not create duplicate demo worlds
- keep seeded credentials easy to use for manual QA

## Core Constraints

- one director belongs to one store only
- stores remain the tenant boundary
- branches remain the operational unit for menu and orders
- menu data stays branch-scoped in the database
- each store owns its own staff, branches, menu, and orders
- secondary stores are included in phase 1 for discovery realism

## Dataset Shape

Phase 1 seeds `3` stores total.

### 1. Demo Bar

Primary store for admin depth and screenshots.

- branches: `6`
- staff: `31`
  - `1` director
  - `6` managers
  - `24` baristas
- users with orders: about `35`
- orders: `120`
- menu categories per branch: `6`
- menu items per branch: about `42`

Suggested branches:

- `Demo Bar - Main`
- `Demo Bar - Downtown`
- `Demo Bar - Riverside`
- `Demo Bar - Chilonzor`
- `Demo Bar - Samarkand Darvoza`
- `Demo Bar - Airport Road`

Suggested category set:

- `Cocktails`
- `Beer`
- `Wine`
- `Coffee`
- `Snacks`
- `Sharing Plates`

### 2. Urban Coffee

Secondary store for discovery variety.

- branches: `3`
- staff: `13`
  - `1` director
  - `3` managers
  - `9` baristas
- users with orders: about `12`
- orders: `45`
- menu categories per branch: `5`
- menu items per branch: about `26`

Suggested category set:

- `Espresso Bar`
- `Signature Drinks`
- `Tea`
- `Pastries`
- `Breakfast`

### 3. Street Burger

Secondary store for discovery variety.

- branches: `3`
- staff: `13`
  - `1` director
  - `3` managers
  - `9` baristas
- users with orders: about `12`
- orders: `45`
- menu categories per branch: `5`
- menu items per branch: about `24`

Suggested category set:

- `Burgers`
- `Chicken`
- `Sides`
- `Combos`
- `Drinks`

## Totals

- stores: `3`
- branches: `12`
- staff: `57`
- users: about `59`
- orders: `210`

## Branch Menu Rules

Menu data remains branch-scoped.

That means:

- each branch gets its own category records
- each branch gets its own item records
- each branch gets its own modifier group records
- each branch gets its own modifier records

Within a store, branch menus should be mostly aligned, not random.

Recommended rule:

- branches in the same store should share the same menu structure and nearly the same item set
- smaller branches may have a small number of unavailable items
- do not make branches within one store feel like unrelated concepts

This keeps the seed aligned with the current schema while still feeling operationally realistic.

## Staff Rules

Each store must have a complete and isolated staff hierarchy.

- every store has exactly `1` director
- every branch has `1` manager
- baristas are branch-scoped
- staff codes should remain human-friendly and predictable for QA

Suggested examples:

- Demo Bar director: `admin`
- Demo Bar branch manager examples: `manager-main`, `manager-downtown`
- barista examples: `barista-main-1`, `barista-main-2`

Secondary stores should use different staff codes and passwords should remain consistent unless there is a clear testing reason to vary them.

## Order Shape

Orders should make the admin panel feel active, but not chaotic.

Recommended time distribution:

- spread orders across the last `14` days
- bias activity toward the last `48` hours
- keep a visible active queue at seed time

Recommended live queue for `Demo Bar`:

- `pending`: `4`
- `accepted`: `5`
- `preparing`: `6`
- `ready`: `5`

The remainder should mostly be:

- `picked_up`
- `cancelled`
- `rejected`

Secondary stores should also have small live queues, but lighter than `Demo Bar`.

Recommended realism rules:

- flagship branches should receive more orders than smaller branches
- basket size should vary from `1` to `5` items
- modifier usage should be common but not universal
- ETA should vary roughly from `5` to `25` minutes
- total prices should vary enough for dashboard and order lists to look believable

## Availability and Operational Detail

To avoid an overly perfect seed:

- some items may be unavailable at smaller branches
- a small number of staff may be inactive
- all branches in the primary store should remain active
- at most one secondary-store branch may be inactive if needed for UI coverage

Use this sparingly. The default seeded world should still feel open and usable.

## Idempotency Rules

The seed command must remain safe to rerun.

- stores should be matched by stable codes
- branches should be matched by stable store-and-branch identifiers
- staff should be matched by stable store-scoped staff codes
- branch menu entities should be matched by stable branch-scoped names
- reruns should update seed-owned records instead of creating duplicates

Orders do not need to preserve the exact same synthetic history forever, but reruns should avoid uncontrolled duplication.

Preferred behavior:

- either clear and recreate seed-owned orders deterministically
- or mark and replace seed-owned orders using a stable seed marker strategy

## Verification Targets

After implementation, the seeded dataset should make these screens visibly populated:

- admin dashboard
- admin branches page
- admin staff page
- admin menu page
- admin orders page
- mini app discovery list and map
- mini app branch detail and menu flow

Manual QA should confirm:

- directors only see their own store
- managers and baristas remain branch-scoped
- each store has distinct branches and menu concepts
- branch-scoped menu records work correctly
- order lists show mixed statuses and realistic totals

## Out of Scope

- changing menu ownership from branch to store
- changing permissions or role boundaries
- introducing platform-wide admin behavior
- adding dozens of extra stores purely for discovery density

## Recommended Implementation Order

1. define structured seed fixtures for stores, branches, staff, categories, items, modifier groups, and modifiers
2. add deterministic seeding for multiple stores and branches
3. add branch-scoped menu generation per store concept
4. add seeded users and realistic order generation
5. verify idempotency and manual QA flows
