# Zero Language — 使用教程

Zero 是一门为 AI 浏览场景设计的编程语言，从第一性原理自举。
目前处于 Phase 1（自举编译器），但语言本身已经可用。

---

## 快速开始

### 构建

`ash
cd go
go build -o ../build/zero.exe .
`

### 运行文件

`ash
echo 'print("hello world")' > hello.z
build/zero.exe hello.z
# 输出: hello world
`

### REPL

`ash
build/zero.exe
# Zero v0.1.0 — Type 'exit' to quit.
# >> print(1 + 2)
# 3
# >>
`

## 语法参考

### 注释

`zero
// 行注释，这是唯一支持的注释形式
`

### 字面量

`zero
42        // 整数
-3.14     // 浮点数
"hello"   // 字符串
true      // 布尔
false
nil       // 空值
`

### 变量

`zero
let x = 42
let name = "zero"
let arr = [1, 2, 3]
let map = {"a": 1, "b": 2}
`

### 重新赋值

`zero
x = 100            // 全局变量赋值
m["key"] = "val"   // map 元素赋值
arr[0] = 99        // 数组元素赋值
`

### 函数

`zero
fn add(a, b) {
    return a + b
}

fn greet(name) {
    print("hello", name)
}
`

### 条件分支

`zero
if x > 0 {
    print("positive")
} else if x < 0 {
    print("negative")
} else {
    print("zero")
}
`

### 循环

`zero
let i = 0
while i < 10 {
    print(i)
    i = i + 1
}
`

### 布尔与比较

`zero
==   !=   <   >   <=   >=   &&   ||   !
`

优先级（从低到高）：
1. ||
2. &&
3. == != < > <= >=
4. + -
5. * / %
6. 一元 - !

### 数组

`zero
let arr = [1, 2, 3]
print(arr[0])     // 取值 -> 1
arr[0] = 99       // 赋值
`

### Map

`zero
let m = {"name": "zero"}
print(m["name"])       // 取值
m["version"] = 1       // 赋值
print(m["missing"])    // 不存在的 key -> nil
`

### 字符串操作

`zero
let s = "hello" + " " + "world"  // 拼接
print(len(s))                    // 长度
print(char_at(s, 0))             // 取字符
`

## 内建函数

| 函数 | 参数 | 说明 |
|------|------|------|
| print(...) | 任意数量 | 空格分隔打印 |
| len(v) | string/array/map | 返回长度 |
| 	ype(v) | 任意 | 返回类型名 |
| char_at(s, i) | string, int | 返回第 i 个字符 |
| ead_file(path) | string | 读取文件 |
| write_file(path, content) | string, string | 写入文件 |

## 已知限制

1. **无 break/continue** — 用 flag 变量控制循环
2. **无块级作用域** — 只有全局和函数参数
3. **函数参数最多 4 个**
4. **无字符串切片** — 不支持 str[i:j]
5. **无 struct** — 用 map 替代
6. **无模块/导入系统** — 所有代码在一个文件

## 路线图

| Phase | 内容 | 状态 |
|-------|------|------|
| 0 | Go 编译器（5570 行） | 完成 |
| 1 | 自举 lexer -> parser -> codegen | 进行中 |
| 2 | AI 浏览器引擎 | 规划 |