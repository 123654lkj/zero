# Zero Phase 2 — 待修清单

## 功能状态总览

| 功能 | 状态 | 说明 |
|------|------|------|
| P2-1 Pattern Matching | ✅ 通过 | `enum`/`match`/`DotExpr` 全链路通 |
| P2-2 Actor Model | ✅ 通过 | `channel()`/`send`/`receive`/`spawn` |
| P2-3 AI Functions | ✅ 通过 | `ai fn` + mock backend |
| P2-4 Effect Tracking | ✅ 通过 | `pure fn` + 效果标注解析 |
| P2-5 Image Snapshots | ❌ restore 崩溃 | `snapshot()` 正常，`restore()` stack underflow |
| P2-6 Pipeline `\|>` | ✅ 通过 | `5 \|> double \|> add1` → 11 |

---

## P2-5 Image Snapshots — restore 崩溃

### 现象
```
let x = 10
snapshot()     // ✅ 正常
x = 20
print(x)       // ✅ 输出 20
restore(0)     // 💥 panic: stack underflow
```

### 根因
`restoreSnapshot()` 在 VM 里恢复了 stackTop 和 frameCount，但没有恢复 stack 数组内容和 frame 数据。当前实现只恢复了计数器，没恢复实际状态。

### 修复方向
`SnapshotEntry` 需要存储：
- `stack []value.Value` (copy)
- `stackTop int`
- `frames []Frame` (copy)
- `frameCount int`
- `globals map[string]value.Value` (deep copy)

`restoreSnapshot()` 需要：
1. 清空当前 stack
2. 从 snapshot 复制 stack 内容
3. 恢复 stackTop, frames, frameCount
4. 恢复 globals

---

## 未写的测试

以下测试文件由 patch 规划但未创建：

| 文件 | 内容 |
|------|------|
| `compiler/pipeline_test.go` | Pipeline 端到端测试 |
| `compiler/lexer_pipe_test.go` | `\|>` token 词法测试 |
| `compiler/parser_pipe_test.go` | Pipeline AST 解析测试 |
| `vm/snapshot_test.go` | Snapshot/Restore 测试 |
| `compiler/effect_test.go` | Effect 标注编译测试 |
| `compiler/parser_effect_test.go` | Effect 标注解析测试 |

---

## 代码统计

| 文件 | 行数 |
|------|------|
| compiler/compiler.go | 823 |
| vm/vm.go | 929 |
| compiler/parser.go | 787 |
| compiler/token.go | 453 |
| opcode/opcode.go | 354 |
| value/value.go | 368 |
| vm/builtins.go | 143 |
| **总代码** | **6942** |

---

## 下一步

1. 修 P2-5 restoreSnapshot — 深拷贝 VM 全状态
2. 补全 6 个测试文件
3. 全量回归测试
4. git commit Phase 2
