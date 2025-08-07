# Wails 最佳实践

## 目录
- [窗体工具栏最大化按钮问题](#窗体工具栏最大化按钮问题)
- [窗口配置最佳实践](#窗口配置最佳实践)
- [macOS特定配置](#macos特定配置)

---

## 窗体工具栏最大化按钮问题

### 问题描述
在Wails应用程序中，系统默认的窗口控制按钮（关闭、最小化、最大化）中的最大化按钮可能显示为灰色，无法点击使用。

### 问题现象
- 最大化按钮（绿色圆圈）显示为灰色或禁用状态
- 点击最大化按钮无响应
- 日志显示函数被调用但窗口状态无变化

### 根本原因分析

#### 1. **窗口尺寸配置缺失**
```go
// ❌ 错误配置 - 缺少关键尺寸设置
err := wails.Run(&options.App{
    Title: "app-name",
    Frameless: false,
    DisableResize: false,
    // 缺少 Width, Height, MinWidth, MinHeight, MaxWidth, MaxHeight
})
```

#### 2. **macOS标题栏配置问题**
```go
// ❌ 错误配置 - 缺少macOS特定配置
err := wails.Run(&options.App{
    Title: "app-name",
    // 缺少 Mac 配置
})
```

#### 3. **无边框模式限制**
```go
// ❌ 错误配置 - 无边框模式可能影响最大化功能
err := wails.Run(&options.App{
    Title: "app-name",
    Frameless: true, // 可能阻止最大化功能
})
```

### 解决方案

#### 1. **完整的窗口尺寸配置**
```go
// ✅ 正确配置 - 包含所有必要的窗口尺寸设置
err := wails.Run(&options.App{
    Title:            "app-name",
    Width:            1024,        // 初始宽度
    Height:           768,         // 初始高度
    MinWidth:         800,         // 最小宽度
    MinHeight:        600,         // 最小高度
    MaxWidth:         0,           // 最大宽度（0表示无限制）
    MaxHeight:        0,           // 最大高度（0表示无限制）
    Frameless:        false,       // 使用标准窗口框架
    DisableResize:    false,       // 允许调整大小
    WindowStartState: options.Normal, // 正常启动状态
})
```

#### 2. **macOS特定配置**
```go
// ✅ 正确配置 - 添加macOS标题栏配置
import "github.com/wailsapp/wails/v2/pkg/options/mac"

err := wails.Run(&options.App{
    Title: "app-name",
    // ... 其他配置
    Mac: &mac.Options{
        TitleBar: &mac.TitleBar{
            TitlebarAppearsTransparent: false, // 标题栏不透明
            HideTitle:                  false, // 不隐藏标题
            HideTitleBar:               false, // 不隐藏标题栏
            FullSizeContent:            false, // 不使用全尺寸内容
            UseToolbar:                 false, // 不使用工具栏
            HideToolbarSeparator:       false, // 不隐藏工具栏分隔符
        },
    },
})
```

#### 3. **运行时窗口控制**
```go
// ✅ 在startup中设置窗口属性
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    
    // 设置窗口尺寸和位置
    runtime.WindowSetSize(ctx, 1200, 800)
    runtime.WindowCenter(ctx)
    
    // 初始化其他功能...
}
```

### 最佳实践总结

#### 1. **始终设置完整的窗口尺寸**
- 设置初始尺寸（Width, Height）
- 设置最小尺寸（MinWidth, MinHeight）
- 设置最大尺寸（MaxWidth, MaxHeight）

#### 2. **平台特定配置**
- macOS：添加Mac配置
- Windows：添加Windows配置（如需要）
- Linux：添加Linux配置（如需要）

#### 3. **避免无边框模式**
- 除非特殊需求，否则使用`Frameless: false`
- 无边框模式可能影响窗口控制功能

#### 4. **测试窗口功能**
```javascript
// 前端测试脚本
async function testWindowState() {
    try {
        const size = await window.runtime.WindowGetSize();
        const isMaximised = await window.runtime.WindowIsMaximised();
        const isFullscreen = await window.runtime.WindowIsFullscreen();
        
        console.log('窗口尺寸:', size);
        console.log('是否最大化:', isMaximised);
        console.log('是否全屏:', isFullscreen);
    } catch (error) {
        console.error('窗口状态检查失败:', error);
    }
}
```

### 常见错误和解决方案

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| 最大化按钮灰色 | 缺少窗口尺寸配置 | 添加Width, Height, MinWidth, MinHeight, MaxWidth, MaxHeight |
| 最大化按钮无响应 | macOS标题栏配置问题 | 添加Mac.TitleBar配置 |
| 窗口无法调整大小 | DisableResize设置为true | 设置DisableResize: false |
| 启动时窗口异常 | WindowStartState配置错误 | 使用options.Normal |

### 调试技巧

1. **检查控制台日志**
   - 查看是否有窗口相关的错误信息
   - 检查API调用是否成功

2. **使用开发模式**
   ```bash
   wails dev
   ```
   - 开发模式提供更多调试信息

3. **测试不同配置**
   - 逐步添加配置项
   - 对比不同配置的效果

4. **检查Wails版本**
   ```bash
   wails version
   ```
   - 确保使用兼容的Wails版本

---

## 窗口配置最佳实践

### 基本配置模板
```go
err := wails.Run(&options.App{
    Title:            "应用名称",
    Width:            1024,
    Height:           768,
    MinWidth:         800,
    MinHeight:        600,
    MaxWidth:         0,
    MaxHeight:        0,
    Frameless:        false,
    DisableResize:    false,
    WindowStartState: options.Normal,
    BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
    
    // 平台特定配置
    Mac: &mac.Options{
        TitleBar: &mac.TitleBar{
            TitlebarAppearsTransparent: false,
            HideTitle:                  false,
            HideTitleBar:               false,
            FullSizeContent:            false,
            UseToolbar:                 false,
            HideToolbarSeparator:       false,
        },
    },
    
    // 其他配置...
})
```

---

## macOS特定配置

### 标题栏配置详解
```go
Mac: &mac.Options{
    TitleBar: &mac.TitleBar{
        // 标题栏外观
        TitlebarAppearsTransparent: false, // 标题栏是否透明
        HideTitle:                  false, // 是否隐藏标题
        HideTitleBar:               false, // 是否隐藏整个标题栏
        
        // 内容区域
        FullSizeContent:            false, // 内容是否占满整个窗口
        
        // 工具栏
        UseToolbar:                 false, // 是否使用工具栏
        HideToolbarSeparator:       false, // 是否隐藏工具栏分隔符
    },
}
```

### 配置说明
- **TitlebarAppearsTransparent**: 设置为false确保标题栏可见
- **HideTitle**: 设置为false确保标题显示
- **HideTitleBar**: 设置为false确保标题栏功能正常
- **FullSizeContent**: 设置为false避免内容区域冲突
- **UseToolbar**: 设置为false使用标准标题栏
- **HideToolbarSeparator**: 设置为false保持标准外观

---

*最后更新: 2025年8月8日*
