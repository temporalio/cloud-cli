# Code Style

## Anchor Comments

Add specially formatted comments throughout the codebase, where appropriate, for yourself as inline knowledge that can be easily `grep`ped for.

- Use `AIDEV-NOTE:`, `AIDEV-TODO:`, or `AIDEV-QUESTION:` (all-caps prefix) for comments aimed at AI and developers.
- **Important:** Before scanning files, always first try to **grep for existing anchors** `AIDEV-*` in relevant subdirectories.
- **Update relevant anchors** when modifying associated code.
- **Do not remove `AIDEV-NOTE`s** without explicit human instruction.
- Make sure to add relevant anchor comments whenever a file or piece of code is:
  - too complex
  - very important
  - confusing
  - could have a bug

## General Rules

- Keep all `const`, `type`, and `var` at the top of the file.
- Group multiple `const`, `type`, and `var` declarations together:
  ```go
  type (
      CompanyName string
      CompanyID   string
  )
  ```
- Limit lines to 120 characters.
- Add an EOF newline for new files you create.

## Avoid Stuttering in Names

When a name's context (the enclosing type or package) already implies part of the name, do not repeat it.

**Bad:**
```go
type User struct {
    UserName string
    UserID   int
}

func (u User) GetUserName() string {
    return u.UserName
}
```

**Good:**
```go
type User struct {
    Name string
    ID   int
}

func (u User) GetName() string {
    return u.Name
}
```
