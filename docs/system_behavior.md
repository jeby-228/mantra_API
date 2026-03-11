# 系統行為流程文件

## 資料表關係圖

```
User (1) ─────┬──── (N) Mantra ──── (N) MantraRecord
              │                  │
              │                  └── (N) MantraDailyStat
              │
              ├──── (N) QuoteRecord ──── (N) MessageBoard
```

---

## 口頭禪模組

### 1. 新增口頭禪 (Mantra)

**觸發條件：** 用戶建立一個新的口頭禪

**流程：**

1. 驗證 `content` 不為空
2. 建立 Mantra 記錄
3. 自動填入 Base 審計欄位 (CreationTime, CreatorId)

**SQL 範例：**

```sql
INSERT INTO mantra (content, description, creation_time, creator_id)
VALUES ('好喔', '常用語', NOW(), 1);
```

---

### 2. 記錄口頭禪出現 (MantraRecord)

**觸發條件：** 用戶記錄某個口頭禪被說出

**流程：**

1. 驗證 `mantra_id` 存在
2. 建立 MantraRecord 記錄
3. 若 `said_at` 為空，使用 `CreationTime` 作為時間
4. **同步更新 MantraDailyStat** (見下方)

**SQL 範例：**

```sql
INSERT INTO mantra_record (mantra_id, location, said_at, creation_time, creator_id)
VALUES (1, '辦公室', '2026-03-11 09:30:00', NOW(), 1);
```

---

### 3. 更新每日統計 (MantraDailyStat)

**觸發條件：** 當 MantraRecord 新增/刪除時自動觸發

**新增 MantraRecord 時：**

```sql
-- 使用 UPSERT：存在則 +1，不存在則建立
INSERT INTO mantra_daily_stat (mantra_id, stat_date, count, creation_time)
VALUES (1, '2026-03-11', 1, NOW())
ON CONFLICT (mantra_id, stat_date)
DO UPDATE SET count = mantra_daily_stat.count + 1,
              last_modification_time = NOW();
```

**刪除 MantraRecord 時：**

```sql
UPDATE mantra_daily_stat
SET count = count - 1, last_modification_time = NOW()
WHERE mantra_id = 1 AND stat_date = '2026-03-11';
```

---

### 4. 查詢每日統計

**用途：** 顯示口頭禪趨勢圖表

```sql
-- 查詢某口頭禪最近 7 天統計
SELECT stat_date, count
FROM mantra_daily_stat
WHERE mantra_id = 1
  AND stat_date >= CURRENT_DATE - INTERVAL '7 days'
ORDER BY stat_date;

-- 查詢所有口頭禪當月總計
SELECT m.content, SUM(s.count) as total
FROM mantra_daily_stat s
JOIN mantra m ON m.id = s.mantra_id
WHERE s.stat_date >= DATE_TRUNC('month', CURRENT_DATE)
GROUP BY m.id, m.content
ORDER BY total DESC;
```

---

## 名言模組

### 5. 新增名言 (QuoteRecord)

**觸發條件：** 用戶記錄一句名言

**流程：**

1. 驗證 `quote` 不為空
2. 建立 QuoteRecord 記錄
3. `said_at` 記錄第一次出現的時間

**SQL 範例：**

```sql
INSERT INTO quote_record (jb_name, quote, said_at, creation_time, creator_id)
VALUES ('小明', '今天也要加油', '2026-03-11 10:00:00', NOW(), 1);
```

---

## 留言板模組

### 6. 新增留言 (MessageBoard)

**觸發條件：** 用戶對某則名言留言

**流程：**

1. 驗證 `quote_record_id` 存在
2. 驗證 `message` 不為空
3. 建立 MessageBoard 記錄
4. `is_edited` 預設為 `false`

**SQL 範例：**

```sql
INSERT INTO message_board (message, quote_record_id, is_edited, creation_time, creator_id)
VALUES ('這句話太經典了!', 1, false, NOW(), 1);
```

---

### 7. 編輯留言

**觸發條件：** 用戶修改自己的留言

**流程：**

1. 驗證留言存在且為本人建立
2. 更新 `message` 內容
3. 將 `is_edited` 設為 `true`
4. 更新 `last_modification_time`

**SQL 範例：**

```sql
UPDATE message_board
SET message = '修改後的內容',
    is_edited = true,
    last_modification_time = NOW(),
    last_modifier_id = 1
WHERE id = 1;
```

---

## Go 程式碼實作範例

### 記錄口頭禪 + 更新統計 (事務處理)

```go
func CreateMantraRecord(db *gorm.DB, record *MantraRecord) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // 1. 建立 MantraRecord
        if err := tx.Create(record).Error; err != nil {
            return err
        }

        // 2. 取得統計日期 (若 said_at 為空則用 CreationTime)
        statDate := record.SaidAt
        if statDate.IsZero() {
            statDate = record.CreationTime
        }
        dateOnly := statDate.Format("2006-01-02")

        // 3. Upsert MantraDailyStat
        result := tx.Exec(`
            INSERT INTO mantra_daily_stat (mantra_id, stat_date, count, creation_time)
            VALUES (?, ?, 1, NOW())
            ON CONFLICT (mantra_id, stat_date)
            DO UPDATE SET count = mantra_daily_stat.count + 1,
                          last_modification_time = NOW()
        `, record.MantraID, dateOnly)

        return result.Error
    })
}
```

---

## 刪除行為 (軟刪除)

所有資料表都使用軟刪除：

- 設定 `is_deleted = true`
- 設定 `deleted_at = NOW()`
- 資料不會真正從資料庫移除

```go
func SoftDelete(db *gorm.DB, model interface{}, id uint) error {
    return db.Model(model).Where("id = ?", id).Updates(map[string]interface{}{
        "is_deleted": true,
        "deleted_at": time.Now(),
    }).Error
}
```
