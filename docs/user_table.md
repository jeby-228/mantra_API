talbe name : Base
| Name | Type | Description |
| ------------- | ------ | ---------------------------------- |
| CreationTime | time.Time | 記錄創建時間 |
| CreatorId | uint | 記錄創建者的用戶ID |
|Sort | int | 排序欄位 (非必填) |
|LastModificationTime | time.Time | 記錄最後修改時間 (非必填) |
|LastModifierId | uint | 記錄最後修改者的用戶ID (非必填) |
|IsDeleted | bool | 軟刪除標記，默認為 false |
|DeleterAt | uint | 記錄刪除者的用戶ID (非必填) |

talbe name : user
| Name | Type | Description |
| ------------- | ------ | ---------------------------------- |
| id | GUID | Primary key, auto-increment |
| username | string | Unique username for the member |
| email | string | Unique email address of the member |
| password_hash | string | Hashed password for authentication |

// 口頭禪主檔
table name : Mantra
| Name | Type | Description |
| --------- | ------ | -------------------------------- |
| id | uint | Primary key, auto-increment |
| content | string | 口頭禪內容 |
| Description | string | 備註欄位 (非必填) |
| + Base | | 審計欄位 |

// 口頭禪紀錄檔
table name : MantraRecord
| Name | Type | Description |
| --------- | --------- | ----------------------------------------------------------------- |
| id | uint | Primary key, auto-increment |
| mantra_id | uint | 外鍵，關聯到 Mantra |
| location | string | 地點 (非必填) |
| said_at | time.Time | 說出口頭禪的時間 (非必填) 如果這邊是空就用 Base 的 CreationTime |
| + Base | | 審計欄位 |

// 口頭禪每日統計
table name : MantraDailyStat
| Name | Type | Description |
| --------- | ---- | ---------------------------------------- |
| id | uint | Primary key, auto-increment |
| mantra_id | uint | 外鍵，關聯到 Mantra |
| stat_date | date | 統計日期 (唯一索引: mantra_id + stat_date) |
| count | int | 當日出現次數 |
| + Base | | 審計欄位 |

// 名言紀錄檔案
table name : QuoteRecord
| Name | Type | Description |
| --------- | --------- | --------------------------- |
| id | uint | Primary key, auto-increment |
| JB_Name | string | 暱稱 姓名 |
| quote | string | 名言內容 |
| said_at | time.Time | 第一次出現的時間 |
| + Base | | 審計欄位 |

// 留言板
table name : MessageBoard
| Name | Type | Description |
| --------- | --------- | --------------------------- |
| id | uint | Primary key, auto-increment |
| message | string | 留言內容 |
| QuoteRecord_id | uint | 外鍵，關聯到 QuoteRecord |
| Isedit | bool | 編輯欄位 |
| + Base | | 審計欄位 |
