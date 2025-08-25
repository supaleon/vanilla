# Vanilla Component File Specification
Vanilla uses HTML, CSS, and Javascript to build UIs. Unlike traditional methods, Vanilla introduces a component system.
Each Vanilla component can be viewed as an HTML file; Vanilla simply enhances the HTML syntax and sets certain limitations.

## File Naming and Location
Vanilla components adopt the standard `.html` extension for HTML files.

Vanilla component files must be stored within the `pages/` folder in the project's root directory and must be named using UpperCamelCase, for example: `Button.html`, `SideBar.html`.
Otherwise, these files will be ignored by the `Vanilla Compiler` and `Vanilla IDE Extension`.

## Component Content Layout

Each component must and can only contain two top-level code blocks:

`Javascript` code block: Enclosed in `<script>` tags, located at the top of the file. It must use `ES Module` syntax, and strict mode is enabled by default.

`Template` code block: Enclosed by a block-level tag (such as `<div>` or `<head>`), or using `<metadata>` as an empty wrapper node.

Apart from these two code blocks, no other top-level tags or comments are allowed in the component file.

No inline `<script>` tags are allowed other than the top-level `<script>`, but the `<script src="..."></script>` form is permitted.

Component Content Layout Examples:
`Example01.html`
```HTML
<script>
    // ES Module code block
</script>

<div>
    <!--Template code block-->
    <span>
        Hello World.
    </span>
    <script src="external.js"></script>
</div>
```

## Component Imports

Use standard ES Module import syntax to bring in other components or stylesheets.

Two types of imports are supported:

**Bare Import** (for side effects): `import "./Item.html"`. Also known as `effect import`.

**Default import**: `import Card from "./Card.html"`. Default imports are mainly used to resolve component name conflicts.

Example:
```html
<script>
    import "./Item.html"
    import Card from "./Card.html"
</script>

<div>
    <Card>
        <Item>Hello World</Item>
    </Card>
</div>
```

## Component Properties (Props)
In Vanilla, a component's property (prop) is a special kind of variable that is processed at compile time. Each prop is either an instance of a Go type or a JS literal.

Properties are declared using a special `prop` macro that is expanded and erased during compilation. This is different from standard runtime variables.

### Property Declaration

Vanilla introduces a `prop` macro within the `Javascript code block`. Component properties are declared using this macro.

Property declarations in Vanilla components must adhere to the following rules:
1. The `prop macro` only accepts instances of Go's composite types (Struct, Map, Slice) or instances of JS literal types, such as boolean, number, string, an empty array (aka `[]`), or an empty object (aka `{}`).
2. Property declarations must use the `const` keyword.
3. Property declarations do not support JS destructuring assignment syntax.

Example:
```HTML
<script>
    import {User} from "./user.go"
    // Equals Go's User type instance `User{}`.
    const user = prop(User())
    // Equals Go's []any type instance `make([]any, 0)`.
    const tags = prop([])
    // Equals Go's string type instance `dark`.
    const theme = prop("dark")
</script>
```

### Property Access

Vanilla supports standard property access (e.g., `{user.disabled}`) and index access (e.g., `{user.tags[0]}`) for member variables within the `Template code block`.
However, it does not support bracket notation access (e.g., `{user["myKey"]}`). This restriction is in place to avoid frequent conflicts with HTML attribute quoting, as attribute values are typically enclosed in double (`"`) or single (`'`) quotes. Furthermore, single quotes denote runes in Go's syntax, which could lead to ambiguity during parsing.

Vanilla also supports a simple boolean unary expression, like `{!user.active}`.

Example:
```HTML
<script>
    import {User} from "./user.go"
    let user = prop(User())
</script>

<div>
    <Card>
        <Item>name: {user.name}</Item>
    </Card>
    <button disabled={!user.active}></button>
</div>
```

### Property Assignment
In Vanilla, if a component is a top-level component, its props must be assigned by the template renderer (which runs in the Go runtime).
In other cases, component props can be assigned similarly to HTML attributes, or through a built-in `<context>` component for cross-level assignment.

`Top.html`
```HTML
<script>
    let theme = prop("dark")
</script>

<div>
    <context theme={theme}>
        <Parent/>
    </context>
</div>
```

## Template Syntax

### Control Flow Directives

#### Conditional Statements
`Vanilla` components support `if` expressions with the following limitations:
1. Only logical and comparison expressions are allowed within an `if` statement, for example: `{if !user.disabled && user.likes > 0}`.
2. `if` statements can have an `else` branch but do not support `else if`.
3. Go's `String` type can be used in `if` statements (e.g., `{if user.code == "NICE"}`). However, Go's raw strings (using backticks ``) are not supported. This is to prevent ambiguity with JavaScript's template literals and to avoid parsing conflicts with HTML, as raw strings do not permit escaping characters (e.g., an expression like `{if user.code == `<a`}` would conflict with HTML tags).
4. Logical and comparison operators must be surrounded by at least one space (e.g., `user.likes > 0` instead of `user.likes>0`). This is to prevent parsing conflicts with HTML tags (e.g., `1<a` could be misinterpreted as the start of an `<a>` tag).

Example:
```html
{if !user.disabled}
    <span>{user.name}</span>
{else}
    <button>Sign In</button>
{/if}
```

#### Loop Statements
Components support `for` loops to iterate over collections or numerical ranges, such as:
`{for index, value in user.tags}` or `{for i, v in 1..9}`

Example:
```html
{for index, value in user.tags}
    <span data-i={index}>{value}</span>
{/for}
```

### Built-in Functions
Vanilla components support a few essential built-in functions within the template. Nested functions are not supported.

- `len(collection)`: Returns the length of a collection (like a slice or map).
- `escape(string)`: Escapes a string for safe rendering in HTML.

Example:
```html
<div>
    {if len(user.tags) > 0}
        User Tags: {for _, tag in tags}<span>{escape(tag)}</span>{/for}
    {/if}
</div>
```

## Stylesheets
Vanilla supports importing CSS stylesheets in components via an import statement, like `import "./style.css"`.
Note that CSS imported by a component is effective in the global context.

Example:
```HTML
<script>
    import "./style.css"
</script>

<div>
    <Card>
        Hello World.
    </Card>
</div>
```