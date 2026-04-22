# 项目分析报告 - golang-zkteco-attshifts

## 1. 业务逻辑概述
本项目是一个基于 Go 语言开发的 ZKTeco（中控）考勤系统数据查询与报表工具。它直接从 ZKTeco 的 SQL Server 数据库中读取考勤记录（attshifts 表）并生成 Web 报表。

## 2. 核心模块说明
- **`internal/service/`**: 数据库交互层。
    - `QueryAtt`: 查询员工每日的出勤、加班、迟到、早退等汇总数据。
    - `QueryLeaveSymbols`: 查询请假标志和异常。
    - `QueryUsersFiltered`: 获取员工及部门信息。
- **`internal/web/`**: 业务逻辑与展现层。
    - `model.go`: 核心报表逻辑，负责按月计算每日出勤、请假天数、各类假期汇总。
    - `render.go` / `grid.go`: 负责将计算出的模型渲染成 HTML 表格、CSV 或 XLS。
    - `columns.go`: 定义了报表中显示的列及其计算方式。

## 3. 请假数据读取逻辑
请假数据主要通过 `internal/service/service.go` 中的 `QueryLeaveSymbols` 函数读取：
- **数据表**: `attshifts`
- **判定条件**: `exceptionid IS NOT NULL AND symbol IS NOT NULL`
- **关键字段**: 
    - `symbol`: 记录请假的符号（通常包含请假时长或类型）。
    - `exceptionid`: 记录假期的类型 ID。
- **类型映射** (在 `internal/web/model.go` 中定义):
    - `case 1`: 公出 (E1Business)
    - `case 2`: 病假 (E2Sick)
    - `case 3`: 事假 (E3Personal)
    - `case 4`: 探亲假 (E4Home)
    - `case 5`: 年假 (E5Annual)

## 4. 已修复问题
- **日期截断问题**: 已修复月份最后一天的查询问题。原逻辑中 `lastDay` 为 00:00:00，导致 `BETWEEN` 查询漏掉最后一天带时间戳的数据。现已将 `lastDay` 调整为该月的 23:59:59。

