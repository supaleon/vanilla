# Vanilla 组件文件规范
Vanilla 使用 HTML、CSS 和 Javascript 来构建 UI，不同于传统的是，Vanilla 引入了组件系统。
每个 Vanilla 组件都可以看成是一个 HTML 文件，Vanilla 只是对 HTML 语法进行了增强，并设定了限制。

## 文件命名与位置
Vanilla 组件沿用了标准 HTML 类型文件的扩展名`.html`。

Vanilla 组件文件必须存放在项目根目录下的 `pages/` 文件夹内，并且使用大驼峰命名，例如：`Button.html`、`SideBar.html`。
否则 `Vanilla Compiler` 和 `Vanilla IDE Extension` 会忽略这些文件。

## 组件内容布局

每个组件必须且只能包含两个顶层代码块：

`Javascript` 代码块：由 `<script> `包裹，位于文件顶部，必须使用 `ES Module` 语法，默认启用严格模式。

`Template` 代码块：由一个块级标签包裹（如 `<div>`、`<head>`），或使用 `<metadata>` 作为空包裹节点。

除这两个代码块外，组件文件中不能有其他顶层标签或顶层注释。

除顶层 `<script>` 外，不允许其他内联 `<script>`，但允许 `<script src="..."></script>` 形式。

## 组件导入

使用标准 ES Module import 语法引入其他组件或样式。

支持两种导入方式：

**Bare Import (用于副作用)**：`import "./Item.html"`。也称`effect import`。

**默认导入**（default import）：`import Card from "./Card.html"`。默认导入主要用于组件名冲突时。

示例：
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

## 组件属性 (Props)
在 Vanilla 中，组件的属性（Prop）是一种在编译时处理的特殊变量。每个属性要么是 Go 类型的实例，要么是 JS 字面量。

属性是使用一个特殊的 `prop` 宏来声明的，它会在编译期间被展开和擦除，这与标准的运行时变量不同。

### 属性声明

Vanilla 在 `Javascript 代码块` 中置入了一个 `prop` 宏。组件属性通过这个宏来声明。

Vanilla 组件的属性声明必须符合以下规范：
1. `prop` 宏接受的参数只能是 Go 的复合类型（Struct、Map、Slice）的实例，或 JS 字面量类型实例，比如 boolean, number, string, empty array (aka `[]`), empty object (aka `{}`).
2. 属性声明必须使用 `const` 关键字。
3. 属性声明不支持 JS 的解构赋值语法。

示例：
```HTML
<script>
    import {User} from "./user.go"
    // 等同于 Go 的 User 类型实例 `User{}`.
    const user = prop(User())
    // 等同于 Go 的 []any 类型实例 `make([]any, 0)`.
    const tags = prop([])
    // 等同于 Go 的 string 类型实例 `dark`.
    const theme = prop("dark")
</script>
```

### 属性访问

Vanilla 支持在 `Template 代码块` 中使用标准的属性访问（如 `{user.disabled}`）和索引访问（如 `{user.tags[0]}`）方式去访问属性变量的成员。
但不支持方括号表示法访问（如 `{user["myKey"]}`）。不支持此语法是为了避免与 HTML 属性的引号冲突，因为属性值通常由双引号（`"`）或单引号（`'`）包裹。此外，单引号在 Go 语言的语法中用于表示 rune 字符，这可能在解析时引起歧义。

Vanilla 同样支持一个简单的布尔一元表达式访问，比如 `{!user.active}`。

示例：
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

### 属性赋值
在 Vanilla 中，如果一个组件是顶层组件，那么它的属性赋值一定来自模版渲染器（渲染器运行在 Go 的运行时中）。
其他情况下，组件属性可以通过类似 HTML Attribute 的方式进行赋值，或通过一个内置的 `<context>` 组件来跨越层级赋值。

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

## 模板语法

### 控制流指令

#### 条件语句
`Vanilla` 组件中支持使用 `if` 表达式，但具有以下限制：
1. `if` 语句内仅允许逻辑表达式与比较表达式，例如：`{if !user.disabled && user.likes > 0}`。
2. `if` 语句可以有 `else` 分支，但不支持 `else if`。
3. `if` 语句中可以使用 Go 的 `String` 类型（例如：`{if user.code == "NICE"}`）。但不支持使用反引号（``）的 `Raw String`。这是为了防止与 JavaScript 的模板字符串产生语法歧义，同时也是为了避免 HTML 解析冲突，因为 `Raw String` 不支持转义，像 `{if user.code == `<a`}` 这样的表达式会与 HTML 标签冲突。
4. 逻辑和比较运算符的两侧必须至少包含一个空格（例如，应使用 `user.likes > 0` 而不是 `user.likes>0`）。这是为了防止与 HTML 标签产生解析冲突（例如，`1<a` 可能会被误解为 `<a>` 标签的开始）。

示例：
```html
{if !user.disabled}
    <span>{user.name}</span>
{else}
    <button>Sign In</button>
{/if}
```

#### 循环语句
组件支持 `for` 循环，可遍历集合或数值区间，如：
`{for index, value in user.tags}` 或 `{for i, v in 1..9}`

示例：
```html
{for index, value in user.tags}
    <span data-i={index}>{value}</span>
{/for}
```

### 内置函数
Vanilla 组件支持在模板中使用几个必须的内置函数。函数不支持嵌套。

- `len(collection)`: 返回一个集合（如 slice 或 map）的长度。
- `escape(string)`: 对字符串进行转义，以便在 HTML 中安全地呈现。

示例：
```html
<div>
    {if len(user.tags) > 0}
        User Tags: {for _, tag in tags}<span>{escape(tag)}</span>{/for}
    {/if}
</div>
```

## 样式表
Vanilla 支持在组件中通过 import 语句导入 CSS 样式表。 如 `import "./style.css"`。 
不过需要注意，组件导入 CSS 样式表是在全局上下文中内有效的。

示例：
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
