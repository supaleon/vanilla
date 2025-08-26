# Vanilla 组件规范

## 核心理念

`Vanilla` 使用 `HTML`、`CSS` 和 `Javascript` 来构建 `UI`，并在 `HTML` 基础上构建了组件系统。

不同于其他框架，`Vanilla` 没有引入复杂的模板指令和语法糖，而是尽量保持组件轻量化。

`Vanilla` 推崇将复杂的数据处理逻辑放进 Go 等后端处理程序中，而不是组件模版里，以便让组件更具可读和可维护性。

如果你需要一个具有复杂模板指令和语法糖的框架，`Vanilla` 并不适合你。`Vanilla` 的哲学是：**代码是写给人看的，只是恰好能在机器上运行。**

## 组件基础

每个 Vanilla 组件都是一个合法的 HTML 文件。

### 文件定义

Vanilla 组件文件必须存放在项目根目录下的 `pages/` 文件夹内，并且使用 `PascalCase（大驼峰）`命名，例如：`Button.html`、
`SideBar.html`。
否则 `Vanilla Compiler` 和 `Vanilla IDE Extension` 会忽略这些文件。

### 内容布局

每个组件至少需要包含一个顶层 `Template` 代码块，可选择地包含一个 `Script` 顶层代码块。除 `Script` 和 `Template`
代码块之外，组件文件不允许包含其他任何顶层标签或注释。另外组件内容还需要满足以下规范：

1. 组件内容必须以一个有效的 `Script` 或 `Template` 代码块开始。
2. `Script` 代码块由 `<script> `包裹，`Script` 代码块内只接受具有严格模式的 `ES Module` 语法的代码。
3. `Template`内容必须由块级 HTML 元素包裹，这个元素可以是 `<div>` 这样的块级标签，也可以是 `<metadata>` 标签。
4. 除顶层 `Script` 代码块 外，不允许使用 `<script>` 内联脚本代码。但允许使用类似 `<script src="./jquery.js"></script>`
   形式外联外部脚本代码。
5. 内联`<style>`代码，只能放在`<head>`或`<metadata>`标签(框架自定义的元数据容器标签)中。

**有效的组件示例：**

`ValidExample01.html`

```HTML

<script>
    import { User } from "./user.go"

    const user = prop(User())
</script>

<div class="container">
    <h1>Hello, {user.name}!</h1>
    {if user.isActive}
    <span class="status-active">Online</span>
    {else}
    <span class="status-inactive">Offline</span>
    {/if}
</div>
```

`ValidExample02.html`

```HTML

<script>
    // Script code block
</script>

<html>
<head>
    <!--Template code block-->
    <style>
        body {
            background: black;
        }
    </style>
</head>
<body>
The answer is 42!
</body>
</html>
```

`ValidExample03.html`

```HTML

<div>
    <!--Template code block-->
    <!-- ✅ 有效！Vanilla 允许组件没有顶层 script 代码块。 -->
    <span>Understand the universe!</span>
</div>
```

`ValidExample04.html`

```HTML

<metadata>
    <!--Template code block-->
    <!-- ✅ 有效！Vanilla 允许使用 metadata 标签代替顶层 Template 代码块。 -->
    <title>My First Vanilla Page</title>
    <meta charset="UTF-8"/>
</metadata>
```

**无效的组件示例：**

~~InvalidExample01.html(多个顶层 Script)~~

```HTML
<!-- ❌ 无效！Vanilla 不支持多个顶层 script 代码块。 -->
<script>
    // Script code block
    console.log("hello world 01")
</script>
<script>
    // Script code block
    console.log("hello world 02")
</script>
<div>template code block</div>
```

~~InvalidExample02.html~~

```HTML
<!-- ❌ 无效！Vanilla 不支持多个顶层 template 代码块。 -->
<div>template code block 01</div>
<div>template code block 02</div>
```

~~InvalidExample03.html~~

```HTML
<!--comment-->
<!-- ❌ 无效！Vanilla 组件源码必须以一个有效的Script或Template开始。 -->
<script>
    console.log("hello world 01")
</script>
<div>template code block</div>
```

~~InvalidExample04.html~~

```HTML

<script>
    console.log("hello world 01")
</script>
<!-- ❌ 无效！Vanilla 源文件必须包含一个顶层 template 代码块。 -->
```

### 模块导入

`Vanilla` 的 `<script>` 代码块支持标准的 `ES Module import` 语法来导入 `Go 类型`、JS 模块、组件或 `CSS` 样式表。

#### 类型导入

Vanilla 支持直接从 `.go` 源文件中导入 Go 的复合类型，以便在 `prop()` 宏中使用。

**语法**: `import { TypeName1, TypeName2 } from "./path/to/your/go/file.go"`

**限制与规范**:

类型导入遵循以下严格的限制：

1.  **支持的类型**: 目前只支持导入 `struct`, `map`, 和 `slice` 三种复合类型。
    *   对于 `map` 类型，其键（key）必须是 `string`。
2.  **导出性**: 导入的类型名称必须以大写字母开头，即在 Go 语言中是“导出（exported）”的。
3.  **不支持的特性**: 不支持导入泛型（generic types）或类型别名（type aliases）。
4.  **实例化规则**:
    *   在 `prop()` 宏中使用导入的类型时，必须以函数调用的形式 `TypeName()` 来进行实例化。
    *   实例化时不支持传递参数或使用字面量为字段赋值。例如，`prop(User({Name: "test"}))` 是无效的。

**示例**:

假设在 `account/user.go` 中定义了一个 `User` 结构体：
```go
// account/user.go

package account

type User struct {
	Name string
	Age  int
}

```

然后你可以在组件中导入并使用它，下面的用法是**正确**的：
```html
<script>
    import { User } from "../account/user.go"

    // User() 使用括号语法来实例化，以获取其零值
    const user = prop(User())
</script>

<div>
    <h1>{user.Name}</h1>
    <p>Age: {user.Age}</p>
</div>
```

#### 组件导入

`Vanilla` 支持两种导入方式导入组件：

**Bare Import (用于副作用)**：`import "./Item.html"`。也称`effect import`。

**默认导入**（default import）：`import Card from "./Card.html"`。默认导入用于获取组件的引用，以便在模板中作为标签使用。当需要时，也可以用它来重命名组件以避免冲突（例如
`import MyCard from "./Card.html"`）。

需要注意的是，`Vanilla` 不支持动态导入组件，例如`import("./Child.html")`，这会引发一个编译错误。

**有效的导入示例：**

```html

<script>
    import "./Item.html"
    import Card from "./Card.html"
</script>

<div>
    <Card>
        <Item/>
    </Card>
</div>
```

**无效的导入示例：**

```html

<script>
    <!-- ❌ 无效！Vanilla 不支持动态导入。 -->
    import("./Card.html")
</script>

<div>
    <Card/>
</div>
```

```html

<script>
    <!-- ❌ 无效！Vanilla 组件文件必须大驼峰命名。 -->
    import card from "./card.html"
</script>

<card/>
```

#### 样式表导入

Vanilla 支持在组件中通过 import 语句导入 CSS 样式表。 如 `import "./style.css"`。
⚠️不过需要注意，通过 import 导入的 CSS 样式表会作用于全局，不支持组件级别的作用域隔离。

**示例：**

```HTML

<script>
    import "./style.css"
</script>

<div class="my-cls"></div>
```

## 组件属性 (Props)

在 Vanilla 中，组件的属性（Prop）是一种在编译时处理的特殊变量。
属性使用一个特殊的 `prop` 宏来声明，它会在编译期间被展开和 `Tree-shaking` 擦除，这与标准的运行时变量不同。

### 属性声明

组件的所有属性，都必须在 `<script>` 块的顶层作用域中，通过全局的 `prop()` 宏进行声明。`prop()` 的调用会返回一个默认值，该值在编译期被处理。

#### `prop()` 宏的参数

`prop()` 的参数定义了属性的**类型**和**默认值**。它可以接受两种类型的参数：

1.  **Go 类型实例化**: 通过 `TypeName()` 的形式传入一个导入的 Go 类型，用于声明一个具有复杂结构的属性。其默认值是该类型在 Go 中的零值。
2.  **JavaScript 字面量**: 直接传入一个 JS 字面量。编译器会根据该字面量推断出对应的 Go 类型和默认值。

**JS 字面量与 Go 类型的映射关系:**

| `prop()` 中的 JS 字面量参数       | 推断出的 Go 类型 |
|:---------------------------| :--- |
| `prop("some string")`      | `string` |
| `prop(123)`                | `int` |
| `prop(1.23)`               | `float64` |
| `prop(true)`               | `bool` |
| `prop([])`                 | `[]any` (Slice) |
| `prop({})`                 | `map[string]any` (Map) |

#### 声明规范

1.  **必须使用 `const`**: 属性声明必须使用 `const` 关键字，以强调其在组件内部的只读特性。
2.  **禁止解构**: 不支持使用解构语法来声明属性。
3.  **禁止 `null` 和 `undefined`**: 不允许使用 `prop(null)` 或 `prop(undefined)` 来初始化属性。

*注意：对于 JS 的数组和对象字面量，目前仅支持使用空字面量 `[]` 和 `{}` 来获取默认值，不支持在其中预设值。*

**有效的示例：**

```HTML
<script>
    import { User } from "./user.go"

    // --- 1. 使用 Go 类型实例化 ---
    // 默认值将是 Go 中 User 类型的零值。
    const user = prop(User());

    // --- 2. 使用 JS 字面量 (类型将被自动推断) ---

    // String -> string
    const theme = prop("dark");

    // Number -> int / float64
    const integerCounts = prop(1);
    const floatDiscount = prop(0.8);

    // Boolean -> bool
    const disabled = prop(false);

    // 暂时仅支持空数组字面量
    // Array literal -> []any
    // 默认值是 []
    const emptyTags = prop([]);
    // Object literal -> map[string]any
    // 默认值是 {}
    const emptyConfig = prop({});
    
</script>
```

**无效的示例：**

```HTML

<script>
    import { User } from "./user.go"

    // ❌ 无效！Vanilla 不支持 let 关键字声明。
    let user = prop(User())
    // ❌ 无效！Vanilla 不支持解构语法。
    const {var1, var2, ...rest} = prop({})
    // ❌ 无效！Vanilla 不支持使用 null 初始化属性。
    const var1 = prop(null)
    // ❌ 无效！Vanilla 不支持使用 undefined 初始化属性。
    const var2 = prop(undefined)
    // ❌ 无效！Vanilla 暂时不支持这种复合类型
    const tags = prop(["hello", "world"]);
    // ❌ 无效！Vanilla 暂时不支持这种复合类型
    const config = prop({ a: 1, b: "2" });
</script>
```

### 属性赋值

组件的属性 (`prop`) 在其被渲染时接收赋值。在 `prop()` 中声明的值是该属性的默认值，当没有外部数据传入时，该默认值会生效。

根据组件的使用方式，属性可以通过以下几种方式被赋值：

#### 1. 直接传递 (Direct Passing)

父组件可以像设置 HTML attribute 一样，直接将数据传递给子组件的同名 `prop`。这是最常见的组件间通信方式。

*在下面的示例中，`Profile.html` 正是通过 `<UserCard user={profile.user} />` 语法，将自身的 `profile.user`
数据传递给了 `UserCard.html` 的 `user` 属性。*

#### 2. 上下文传递 (Contextual Passing)

对于跨越多层级的属性传递，可以使用内置的 `<context>` 组件。它可以将其包裹的所有子组件（无论层级多深）的同名 `prop`
进行赋值，从而避免逐层手动传递。

*在下面的示例中，`Profile.html` 的 `theme` 属性值 `dark`，通过 `<context>` 直接传递给了孙代组件 `TagList.html`
，并覆盖了其内部原有的默认值 `light`。*

#### 3. 顶层渲染器注入

如果一个组件是直接被 Go 模板渲染器渲染的顶层页面，那么它的 `prop` 值将直接由渲染器在执行时提供。

**示例：**
`Profile.html`

```HTML

<script>
    import { Profile } from "./profile.go"
    import "./UserCard.html"
    
    const profile = prop(Profile())
    const theme = prop("dark")
</script>

<div>
    <context theme={theme}>
        <UserCard user={profile.user}/>
    </context>
</div>
```

`UserCard.html`

```HTML

<script>
    import { User } from "./user.go"
    import "./TagList.html"
    // Accepts the value from `Profile.html` component.
    const user = prop(User())
</script>

<div>
    <TagList tags={user.tags}/>
</div>
```

`TagList.html`

```HTML

<script>
    import { Tags } from "./tag.go"
    // Accepts the value from `UserCard.html` component.
    const tags = prop(Tags())
    // Will be overwritten by context from `Profile.html` component.
    const theme = prop("light")
</script>

<div class="{theme}">
    <div>
        User Tags:
        {for _, tag in tags}
        <span>{tag}</span>
        {/for}
    </div>
</div>
```

### 属性访问

在模板（`Template` 代码块）中，你可以通过 `{}` 插值语法来访问 `prop` 声明的属性。

#### 支持的语法

* **点表示法 (Dot Notation):** 用于访问对象属性，例如 `{user.name}`。
* **索引访问 (Index Access):** 用于访问数组成员，例如 `{user.tags[0]}`。
* **布尔“非”运算符 (Boolean NOT Operator):** 用于对布尔值取反，例如 `{!user.active}`。

#### 不支持的语法与原因

为了确保模板解析的健壮性和明确性，存在以下限制：

* **不支持方括号表示法 (Bracket Notation):** 完全禁止 `{user["myKey"]}` 这样的语法。
    * **原因**: HTML 属性值本身常用引号 (`"` 或 `'`) 包裹，在插值中再使用引号会引起解析冲突和歧义。此外，Go 语言中单引号用于表示
      `rune` 字符，这也会导致语法的不一致性。
* **不支持复杂的表达式:** 模板插值仅限于直接的属性访问和简单的布尔 `!` 运算，不支持算术运算或其他复杂的表达式。

**有效的示例：**

```HTML

<script>
    import { User } from "./user.go"

    const user = prop(User())
</script>

<div>
    <span>name: {user.name}</span>
    <span>{empty(user.tags): 未设置标签}</span>
    <button disabled={!user.active}></button>
</div>
```

**无效的示例：**

```HTML

<script>
    import { User } from "./user.go"

    const user = prop(User())
</script>

<!--❌ 无效！Vanilla 不支持方括号表示法访问。-->
<div title="{user[" name"]}">
<!--❌ 无效！Vanilla 不支持方括号表示法访问。-->
<button disabled={!user['active']}></button>
</div>
```

## 模板语法

### 条件渲染

`Vanilla` 组件中支持使用 `if` 表达式，但具有以下限制：

1. `if` 语句内仅允许逻辑表达式（||和&&）与比较表达式(>、<、<=、>=)，例如：`{if !user.disabled && user.likes > 0}`。
2. `if` 语句可以有 `else` 分支，但不支持 `else if`。
3. `if` 语句中可以使用 Go 的 `String` 类型（例如：`{if user.code == "NICE"}`）。但不支持使用反引号(``)的 `Raw String`
   。这是为了防止与 JavaScript 的模板字符串产生语法歧义，同时也是为了避免 HTML 解析冲突，因为 `Raw String` 不支持转义，像
   `{if user.code == `<a`}` 这样的表达式会与 HTML 标签冲突。
4. 逻辑和比较运算符的两侧必须至少包含一个空格（例如，应使用 `user.likes > 0` 而不是 `user.likes>0`）。这是为了防止与 HTML
   标签产生解析冲突（例如，`1<a` 可能会被误解为 `<a>` 标签的开始）。

**示例：**

```html

<div>
    {if !user.disabled}
    <span>{user.name}</span>
    {else}
    <button>Sign In</button>
    {/if}
</div>
```

#### 布尔属性表达式

HTML 中的布尔属性（如 `disabled`, `checked`, `selected`）比较特殊，它们的特点是“存在即为 true”。

Vanilla 提供了一种简洁的语法来控制这些属性。在为这类属性赋值时可以使用布尔属性表达式：

* 如果表达式的结果为 `true`，该属性就会被渲染到 HTML 标签上。
* 如果表达式的结果为 `false`，该属性则会被完全移除。

**示例：**

```html

<script>
    const user = prop({})
    // 假设 user = {
    //      isLoggedIn: false,
    //      sDeactivated: true,
    //  }
</script>
<div>
    <!-- 
      因为 !user.isLoggedIn 的结果是 true, 
      所以最终渲染出的 button 会是 <button disabled>。
    -->
    <button disabled={!user.isLoggedIn}>提交</button>

    <!-- 
      因为 user.isDeactivated 的结果是 true, 
      所以最终渲染出的 input 会是 <input type="checkbox" checked>。
    -->
    <input type="checkbox" checked={user.isDeactivated}/>
</div>
```

#### 条件文本表达式

这是一种 `if/then` 的简写形式，特别适合用于动态渲染 CSS 类名。

**语法**: `{ condition: value_if_true }`

`value_if_true` 是一个**不需要引号**的原始字符串，解析器会从 `:` 之后一直读取到 `}` 作为 `true` 条件下的输出值。

*   如果 `condition` 的结果为 `true`，表达式会输出 `value_if_true`。
*   如果 `condition` 的结果为 `false`，表达式会输出一个空字符串。

这种语法可以和静态文本组合使用，轻松实现一个动态的 `class` 列表。

**示例：**

```html

<script>
    const state = prop({})
    // 假设在渲染时被赋值为 state = {
    //  isVip: true,
    //  isDarkMode: false
    // }
</script>
<!-- 
  因为 state.isVip 是 true, 
  class 属性会被渲染为 "card vip-badge"。
-->
<div class="card {state.isVip: vip-badge}">VIP 卡片</div>

<!-- 
  因为 state.isDarkMode 是 false, 
  条件文本表达式部分输出空字符串，
  class 属性最终为 "card"。
-->
<div class="card {state.isDarkMode: dark-mode}">普通卡片</div>

<!-- 
  组合使用：
  因为 state.isVip 是 true, state.isDarkMode 是 false,
  class 属性最终被渲染为 "card vip-badge"。
-->
<div class="card {state.isVip: vip-badge} {state.isDarkMode: dark-mode}">组合卡片</div>
```

### 循环渲染

`Vanilla` 使用 `{for ...}` 和 `{/for}` 块来实现循环渲染，其语法类似于 Go 语言的 `for...range`
语句。循环渲染主要支持两种模式：遍历集合和遍历数值区间。

#### 遍历集合 (Iterating over Collections)

你可以遍历一个数组（Slice）或字典（Map）类型的 `prop`。

**语法**: `{for index, value in collection}`

* `collection`: 你要遍历的 `collection`。
* `index`: 集合中当前元素的索引或键。
* `value`: (可选) 集合中当前元素的值。


**示例**:

组件内容：
```html

<script>
    const tags = prop(["tech", "news", "sports"])
</script>
<ul>
    {for i, tag in tags}
    <li>{i}: {tag}</li>
    {/for}
</ul>
```

渲染结果：

```html

<ul>
    <li>0: tech</li>
    <li>1: news</li>
    <li>2: sports</li>
</ul>
```

#### 遍历数值区间 (Iterating over Numerical Ranges)

你可以使用 `start..end` 的语法来遍历一个闭区间整数，`start` 和 `end` 必须是 `int` 类型常量。

**语法**: `{for i, v in start..end}`

* `start..end`: 一个包含起始值和结束值的闭区间。
* `v`: 当前区间的数值。
* `i`: (可选) 从 0 开始的索引。

**示例**:

组件内容:
```html

<div>
    {for _, num in 1..5}
    <span>{num}</span>
    {/for}
</div>
```

渲染结果：

```html

<div>
    <span>1</span>
    <span>2</span>
    <span>3</span>
    <span>4</span>
    <span>5</span>
</div>
```

### 格式化输出

你可以在模板插值中，使用百分号 `%` 来对变量进行格式化。

**语法**: `{ variable % format_specifier }`

格式化限定符 `format_specifier` 是一个**不需要引号**的原始字符串，解析器会从 `%` 一直读取到 `}` 作为完整的限定符。

格式化规则由变量的类型决定：

* **数字 (Numbers)**: 若变量为数字（整数或浮点数），`format_specifier` 遵循 Go `fmt` 包的格式化动词。

* **时间 (Time)**: 时间格式化支持两种模式，具体取决于 `prop` 的数据类型：
    1. **Go 标准布局**: 如果 `variable` 是一个 `time.Time` 对象，`format_specifier` 遵循 Go `time` 包的布局字符串。
    2. **便捷写法**: 如果 `variable` 是一个 Unix 时间戳（`int64`），`format_specifier` 可以使用更简洁的写法，如 `YY/MM/DD H:M:S`。

**示例**:

组件内容:
```html

<script>
    import { Sale } from "./sale.go"
    
    const sale = prop(Sale())
    // 假设 sale = {
    //      price: 12.345,
    //      dealExpiresAt: someGoTimeObject, // time.Time 对象, 值为 2025-08-25
    //      shippedAt: 1756112902,          // int64 时间戳, 值为 2025-08-25 17:08:22
    // }
</script>
<!-- 数字格式化 -->
<p>Price: ${sale.price % "%.2f"}</p>

<!-- 时间格式化 (time.Time 对象) -->
<p>Expires: {sale.dealExpiresAt % 2006-01-02}</p>

<!-- 时间格式化 (Unix 时间戳) -->
<p>Shipped: {sale.shippedAt % YY/MM/DD H:M:S}</p>
```

**渲染结果**:

```html
<p>Price: $12.35</p>
<p>Expires: 2025-08-25</p>
<p>Shipped: 2025-08-25 17:08:22</p>
```

### 内置函数

Vanilla 在模板中提供了一组内置函数，用于处理常见的格式化和逻辑需求。所有内置函数均不支持嵌套调用（如`len(ok(collection))`）。

* `len(collection)`
  返回集合类型（如 Slice 或 Map等）的长度。

* `ok(path.to.value)`
  安全地检查深度嵌套的属性是否存在。如果访问路径中的任何部分为 `nil`，表达式返回 `false`，可以有效防止渲染时出错。

* `empty(variable)`
  检查变量值是否为空 (`nil`, `null`, `""`)。如果属性路径不存在，同样返回 `true`。

* `lower(string)`
  将字符串转换为全小写。

* `upper(string)`
  将字符串转换为全大写。

* `unsafe(html_string)`
  禁止对变量进行 HTML 转义。默认情况下，所有模板输出都会被自动转义以防止 XSS 攻击。`unsafe` 用于显式地输出原始 HTML。
  **警告**：⚠️仅在完全信任内容来源时使用，否则会带来安全风险。

  **上下文安全特性**:
  为了防止通过 `href` 注入恶意脚本，当 `unsafe` **直接用在 `<a>` 标签的 `href` 属性**中时，它会表现出特殊的安全行为：它依然会移除危险的 URL 协议（如 `javascript:`），但允许 `http`, `https` 等安全协议通过。在其他任何地方，它都只是纯粹地禁止转义。

**`unsafe` 示例**:

组件内容：
```html
<script>
    const links = prop({
        safeUrl: "https://example.com",
        maliciousUrl: "javascript:alert('XSS')",
        rawHtml: "<span>Raw HTML</span>"
    })
</script>
<!-- 在 href 中，危险协议被移除 -->
<a href="{unsafe(links.maliciousUrl)}">Malicious Link</a>

<!-- 在 href 中，安全协议被保留 -->
<a href="{unsafe(links.safeUrl)}">Safe Link</a>

<!-- 在非 href 的地方，完全不转义 -->
<div>{unsafe(links.rawHtml)}</div>
```
渲染结果：
```html
<a href="">Malicious Link</a>
<a href="https://example.com">Safe Link</a>
<div><span>Raw HTML</span></div>
```

**统一示例**:

`post.go`

```go
package post

type Post struct {
	Title  string
	Tags   []string
	Author struct {
		Name string
		// Profile field is missing
	}
	ContentHtml string
	Status      string
}
```

组件内容：

`Post.html`

```html

<script>
    import { Post } from "./post.go"

    const post = prop(Post())
    // 假设 post = {
    //    Title: "My Post",
    //    Tags:  ["tech", "review"],
    //    Author: {
    //        Name: "John Doe",
    //    },
    //    ContentHtml: "<em>Hello</em>",
    //    Status:      "",
    //}
</script>

<!-- len -->
<p>标签数: {len(post.tags)}</p>

<!-- upper -->
<p>标题: {upper(post.title)}</p>

<!-- ok: 安全地访问 post.author.profile.image -->
{if ok(post.author.profile.image)}
<img src="{post.author.profile.image}">
{else}
<p>作者没头像</p>
{/if}

<!-- empty: 检查 status 是不是空字符串 -->
{if empty(post.status)}
<p>状态: 未发布</p>
{/if}

<!-- unsafe: 渲染 HTML 内容 -->
<div>内容: {unsafe(post.contentHtml)}</div>
```

**渲染结果**:

```html
<p>标签数: 2</p>
<p>标题: MY POST</p>
<p>作者没头像</p>
<p>状态: 未发布</p>
<div>内容: <em>Hello</em></div>
```
