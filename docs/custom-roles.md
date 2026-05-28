# Custom Roles

Custom roles let you bind a named set of permissions (resource type + actions)
to a role ID, and then assign that role to users, user groups, and account-
scoped service accounts.

This document covers two things:

1. **Managing custom roles** — `temporal cloud custom-role` (create, update,
   get, list, delete, apply, edit).
2. **Assigning custom roles to principals** — users, user groups, and account-
   scoped service accounts.

A principal (user, user group, or service account) must have a built-in
account role (e.g. `developer`, `read`). Custom roles are an additional
optional set of permissions on top of that account role.

> Namespace-scoped service accounts cannot have custom roles assigned.

---

## 1. Managing custom roles

### The custom role spec

A custom role is defined by a JSON spec:

```json
{
  "name": "namespace-reader",
  "description": "Read-only access to all namespaces",
  "permissions": [
    {
      "resources": {
        "resource_type": "namespace",
        "allow_all": true
      },
      "actions": ["cloud.namespace.get"]
    }
  ]
}
```

- `name` — required, the human-readable name.
- `description` — optional.
- `permissions[]` — list of permission entries:
  - `resources.resource_type` — what the permission applies to.
  - `resources.resource_ids` — specific resource IDs (omit if `allow_all`).
  - `resources.allow_all` — true to apply to every resource of the type within the current parent.
  - `actions[]` — the actions allowed (e.g. `cloud.namespace.get`, `cloud.namespace.update`).

Save the spec to a file (e.g. `role.json`) so it can be passed via `@file`.

### Create a custom role

```bash
temporal cloud custom-role create --spec @role.json
```

Inline JSON also works:

```bash
temporal cloud custom-role create --spec '{"name":"reader","permissions":[{"resources":{"resource_type":"namespace","allow_all":true},"actions":["cloud.namespace.get"]}]}'
```

The command prompts for confirmation, then prints the new role ID.

### Get a custom role

```bash
temporal cloud custom-role get --role-id my-role-id
```

### List custom roles

```bash
temporal cloud custom-role list
temporal cloud custom-role list --page-size 50
```

### Update a custom role

`update` replaces the spec wholesale by ID:

```bash
temporal cloud custom-role update --role-id my-role-id --spec @role.json
```

You can override the resource version explicitly with `--resource-version`;
otherwise the latest version is fetched and used.

### Delete a custom role

```bash
temporal cloud custom-role delete --role-id my-role-id
```

### Apply (create-or-update by name)

`apply` looks up the existing role by `name` and updates it if found, or
creates it if not:

```bash
temporal cloud custom-role apply --spec @role.json
```

This shows a diff between the existing and new spec before prompting.

> If multiple roles share the same name, `apply` errors and asks you to use
> `update --role-id` instead.

### Edit interactively

`edit` opens the spec in `$EDITOR` (falling back to `vi`):

```bash
temporal cloud custom-role edit --role-id my-role-id
```

Save and close to apply; the diff is shown before confirming.

---

## 2. Assigning custom roles to principals

Custom roles attach to a principal's `AccountAccess.CustomRoles` list. The
list is a set of role IDs.

### Pattern: `set-custom-roles` (replace the list)

Each principal type has a `set-custom-roles` command that **replaces** the
current list with whatever you pass. Pass no `--custom-role` flags to remove
all custom roles.

| Principal           | Command                                  |
| ------------------- | ---------------------------------------- |
| User                | `temporal cloud user set-custom-roles`   |
| User group          | `temporal cloud user-group set-custom-roles` |
| Service account     | `temporal cloud service-account set-custom-roles` |

A diff is shown before each call.

### Users

Add or replace custom roles on a user (by email or ID):

```bash
temporal cloud user set-custom-roles \
  --user-email alice@example.com \
  --custom-role role-reader-id \
  --custom-role role-deployer-id
```

Remove all custom roles:

```bash
temporal cloud user set-custom-roles --user-email alice@example.com
```

You can also assign custom roles when inviting a user (requires
`--account-role`):

```bash
temporal cloud user invite \
  --email alice@example.com \
  --account-role developer \
  --custom-role role-reader-id
```

### User groups

```bash
temporal cloud user-group set-custom-roles \
  --group-id my-group-id \
  --custom-role role-reader-id
```

Remove all custom roles:

```bash
temporal cloud user-group set-custom-roles --group-id my-group-id
```

You can assign custom roles when creating a group (requires
`--account-role`):

```bash
temporal cloud user-group create-cloud-group \
  --display-name "Engineering" \
  --account-role developer \
  --custom-role role-reader-id
```

The same `--custom-role` flag works on `create-google-group` and
`create-scim-group`.

`user-group update` also accepts `--custom-role` (replaces the list when
provided; leaves it alone when not):

```bash
temporal cloud user-group update \
  --group-id my-group-id \
  --custom-role role-deployer-id
```

To clear custom roles via the `update` flow, use `set-custom-roles` with
no flags instead.

### Service accounts

Custom roles only apply to **account-scoped** service accounts. They cannot
be set on namespace-scoped service accounts.

```bash
temporal cloud service-account set-custom-roles \
  --service-account-id my-sa-id \
  --custom-role role-reader-id
```

Remove all custom roles:

```bash
temporal cloud service-account set-custom-roles --service-account-id my-sa-id
```

Assign at create time (requires `--account-role`):

```bash
temporal cloud service-account create \
  --name my-sa \
  --account-role developer \
  --custom-role role-reader-id
```

`service-account update` also accepts `--custom-role` for replacing the
list. To clear, use `set-custom-roles` with no flags.

---

## Notes

- **Built-in role required.** A principal must have a built-in
  `--account-role` (e.g. `developer`). `--custom-role` is rejected at
  create/invite time without one.
- **Preserved on role change.** `set-account-role` for a user or user group
  preserves the existing custom role list — you don't have to re-specify it
  when swapping the built-in role.
- **De-duplication.** Repeating `--custom-role role-x --custom-role role-x`
  is silently de-duplicated.
- **Idempotency.** Updates that result in no change are rejected by the
  server; pass `--idempotent` (where the flag exists) to succeed silently
  in that case.
