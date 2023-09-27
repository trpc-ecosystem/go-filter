# tRPC-Go [degrade] 熔断保护插件

## degrade 插件使用介绍

* degrade 插件主要功能：

1. 周期检测获取系统负载情况
2. 根据 CPUidle，内存使用率，负载（主要 load5）来设置阈值，达到阈值触发熔断保护，抛弃一定百分比的随机流量
3. 可限制最大并发请求数，应对超短时突发流量，和上述互为补充

**备注：**

假如平台上的 docker 容器是采用软隔离的方式，那么 top,free 命令看到的数据都是宿主机的数据，常规的采集指标（/proc/(mem|cpu)）也是宿主机数据。获取 Docker 容器的资源利用率需要两项基础技术 Namespace 和 cgroup。

目前可以针对 cpu 利用率和内存利用率这两个指标做熔断操作。

cpu 和内存指标计算方式如下：

一、cpu 使用率计算

平台采用时间片的方式来调度每一个容器的 CPU 使用时间，每个容器在同一个调度周期内可以使用的 cpu 时间不能超过设定的配额。在 cgroup 中用 cpu.cfs_period_us 指定调度周期，用 cpu.cfs_quota_us 来指定相应容器可以使用 cpu 的最大时间长度，一个调度周期时内容器可以使用的最大 cpu 时间是：CPU 配额比机器 cpu 核心数调度周期。

调度周期默认是 100000 微秒。

例如，一个 48 核的宿主机上有一个 25% 配额的容器，这个容器可以使用的最大 cpu 时间是

0.25 * 48 * 100000 = 1200000（微秒）

计算 cpu 配额（容器里按照实际 cgroup 文件中算，其实为超配）

限制 cpu 核数计算：

cpu.cfs_quota_us/cpu.cfs_period_us
cpu 核数：

/sys/fs/cgroup/cpuacct/cpuacct.usage_percpu

cgroup 中 cpuacct.usage 有统计了容器从创建以来所使用的 cpu 时间和。

获取 cpuacct.usage
Sleep interval 间隔纳秒
获取 cpuacct.usage
计算 cpu 使用率：（两次获取 cpuacct.usage 的差值/（interval * 容器 cpu 配额））
二、内存利用率计算

Total: cgroup 被限制可以使用多少内存，可以从文件里的 hierarchical_memory_limit 获得，但不是所有 cgroup 都限制内存，没有限制的话会获得 2^64-1 这样的值，我们还需要从 /proc/meminfo 中获得 MemTotal，取两者最小。
RSS: Resident Set Size 实际物理内存使用量，在 memory/memory.stat 的 rss 只是 anonymous and swap cache memory，文档里也说了如果要获得真正的 RSS 还需要加上 mapped_file。
Cached: memory/memory.stat 中的 cache
MappedFile: memory/memory.stat 中的 mapped_file、

这里不将共享内存计算入容器使用内存

## 使用说明

### load5 参数值设置参考（常规机器/容器）

### 当获取 load5 不准时可以将 load 的阈值设置为 999

由于 5 分钟 load 可能存在对系统负载不敏感的地方，因此根据不断的长期研究和实践，结合在实际业务中的使用情况，建议使用如下配置方法

```go
    load5 := int ((2 * x + 3 * y) / 5)
```

其中 x 为空闲期平均 1 分钟负载，y 为核心数（一般负载超过核心数认为高负载）
理论为 连续 5 分钟周期，3 分钟高负载即触发 熔断！

假如平台容器的 load 指标不能准确获取，需要设置一个大值以跳过这个指标，例如：10000

例如：
使用 uptime 命令：

```go
[xxx@xxx ~]$ uptime
 14:55:45 up 121 days, 15:55,  1 user,  load average: 0.28, 0.26, 0.23
load average: 0.28(1min), 0.26(5min), 0.23(15min)

比如 load average: 1min 负载为 0.16, 核心数为 2, 则 load 5min 配置 load5 参考值设置为 1.264

(0.16*2+3*2)/5=(0.32+6)/5=6.32/5=1.264

核心数计算：grep -c 'model name' /proc/cpuinfo

```

* 增加 import

```go
import (
   _ "trpc.group/trpc-go/trpc-filter/degrade"
)
```

* 增加配置

```yaml
server:
  filter:
    - degrade
plugins:
  circuitbreaker:
    degrade:
      load5: 5000          # load 5 分钟 触发熔断的阈值，恢复时使用实时 load1 <= 本值来判断，更敏感
      cpu_idle: 30         # cpu 空闲率，低于 30%，进入熔断
      memory_use_p : 60    # 内存使用率超过 60%，进入熔断
      degrade_rate : 60    # 流量保留比例，目前使用随机算法抛弃。0 或 100 则不启用此插件
      interval : 30        # 心跳时间间隔，主要控制多久更新一次熔断开关状态，单位"s"
      max_concurrent_cnt : 10000  # 最大并发数
      max_timeout_ms : 100        # 超过最大并发请求数时，最多等待 MaxTimeOutMs 才决定是丢弃还是继续处理
```

字段说明如下：

```go
type Config struct {
    Load5             float64 `yaml:"load5"`              // load 5 分钟 触发熔断的阈值，恢复时使用实时 load1 <= 本值来判断，更敏感
    CPUIdle           int     `yaml:"cpu_idle"`           // cpuidle 触发熔断的阈值
    MemoryUserPercent int     `yaml:"memory_use_p"`       // 内存使用率百分比 触发熔断的阈值
    DegradeRate       int     `yaml:"degrade_rate"`       // 流量保留比例，目前使用随机算法抛弃，迭代加入其他均衡算法
    Interval          int     `yaml:"interval"`           // 心跳时间间隔，主要控制多久更新一次熔断开关状态
    Modulename        string  `yaml:"modulename"`         // 模块名，后续用于上报鹰眼或其他日志
    Whitelist         string  `yaml:"whitelist"`          // 白名单，用于后续跳过不被熔断控制的业务接口
    IsActive          bool    `yaml:"-"`                  // 标志熔断是否生效
    MaxConcurrentCnt  int     `yaml:"max_concurrent_cnt"` // 最大并发请求数，<=0 时不开启。和上述熔断互为补充，能防止突发流量把服务打死，比如 1ms 内突然进入 100W 请求
    MaxTimeOutMs      int     `yaml:"max_timeout_ms"`     // 超过最大并发请求数时，最多等待 MaxTimeOutMs 才决定是丢弃还是继续处理
}
```
