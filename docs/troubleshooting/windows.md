# Windows Troubleshooting (PowerShell Installer)

This doc covers common Windows issues when running the one-line PowerShell installer:

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; irm https://raw.githubusercontent.com/kittors/AgentFlow/main/install.ps1 | iex
```

---

## zh-CN：Windows 常见问题

### 1) 报错：`基础连接已经关闭: 接收时发生错误`

**原因**：Windows PowerShell 5.1 默认不启用 TLS 1.2，而 GitHub 要求 TLS 1.2+。

**解决**：在命令前加上 TLS 设置（上面已包含此修复）：

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; irm https://raw.githubusercontent.com/kittors/AgentFlow/main/install.ps1 | iex
```

如果你使用的是 PowerShell 7+（`pwsh`），则不需要此前缀。

### 2) 报错：`未能解析此远程名称: 'raw.githubusercontent.com'`

**结论**：这是网络/DNS/代理导致的域名解析失败，通常不是脚本本身的问题。

#### 快速定位（建议先做）

在 PowerShell 里执行：

```powershell
Resolve-DnsName raw.githubusercontent.com
nslookup raw.githubusercontent.com
```

如果这里就解析失败，优先按下面的 DNS/代理步骤处理。

#### 方案 A：刷新 DNS 缓存

```powershell
ipconfig /flushdns
```

然后重试安装命令。

#### 方案 B：切换到稳定的 DNS（最常见有效）

先查看当前 DNS：

```powershell
Get-DnsClientServerAddress
```

把网卡 DNS 改为公共 DNS（把 `Wi-Fi` 替换成你自己的网卡名，比如 `Ethernet`）：

```powershell
Set-DnsClientServerAddress -InterfaceAlias "Wi-Fi" -ServerAddresses 1.1.1.1,8.8.8.8
```

改完后重新验证：

```powershell
Resolve-DnsName raw.githubusercontent.com
```

如需恢复为 DHCP 自动分配 DNS：

```powershell
Set-DnsClientServerAddress -InterfaceAlias "Wi-Fi" -ResetServerAddresses
```

#### 方案 C：检查代理（公司网络很常见）

查看 WinHTTP 代理：

```powershell
netsh winhttp show proxy
```

如果你所在网络不需要代理但这里配置了代理，可能会导致请求异常（谨慎执行重置）：

```powershell
netsh winhttp reset proxy
```

如果你必须使用代理（公司/校园网/本机代理软件），请确保系统代理或 `HTTP(S)_PROXY` 配置正确。

#### 方案 D：绕过 `raw.githubusercontent.com`（raw 被拦截时非常实用）

如果 `github.com` 可访问但 `raw.githubusercontent.com` 被拦：

```powershell
winget install Git.Git
git clone https://github.com/kittors/AgentFlow.git
cd AgentFlow
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

或者直接从 GitHub Releases 下载 `agentflow-windows-amd64.exe`，手动放到 `PATH`。

---

## Notes

- `irm` is an alias of `Invoke-RestMethod` in PowerShell.
- If you can open GitHub in a browser but PowerShell still fails, it is often because the browser uses DoH while the OS resolver is still broken or blocked.
