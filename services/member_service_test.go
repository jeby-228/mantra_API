package services

import (
	"testing"

	"mantra_API/audit"
	"mantra_API/auth"
	"mantra_API/internal/testhelper"
	"mantra_API/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var (
	memberCreator1 = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	memberCreator2 = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

// insertMemberRow 直接寫入 DB：CreateMember 未設定 LineID，多筆空字串會違反 line_id unique（SQLite）。
func insertMemberRow(t *testing.T, db *gorm.DB, name, email string, creator uuid.UUID) {
	t.Helper()
	hash, err := auth.HashPassword("pw")
	require.NoError(t, err)
	m := &models.Member{
		Base:         audit.NewCreateBase(creator),
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		LineID:       uuid.New().String(),
	}
	require.NoError(t, db.Create(m).Error)
}

func TestMemberService_CreateMember_Success(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMemberService(db)

	m, err := svc.CreateMember("Alice", "alice@example.com", "secret123", memberCreator1)
	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Equal(t, "Alice", m.Name)
	assert.Equal(t, "alice@example.com", m.Email)
	assert.True(t, auth.CheckPassword("secret123", m.PasswordHash))
	assert.Equal(t, memberCreator1, m.CreatorId)
}

func TestMemberService_CreateMember_DuplicateEmail(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMemberService(db)

	_, err := svc.CreateMember("A", "dup@example.com", "p1", memberCreator1)
	require.NoError(t, err)

	m2, err := svc.CreateMember("B", "dup@example.com", "p2", memberCreator2)
	assert.Error(t, err)
	assert.Nil(t, m2)
	assert.Equal(t, "email 已被使用", err.Error())
}

func TestMemberService_UpdateMember_Success(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMemberService(db)

	created, err := svc.CreateMember("Old", "old@example.com", "pw", memberCreator1)
	require.NoError(t, err)

	updated, err := svc.UpdateMember(created.ID, "New", "new@example.com", memberCreator2)
	require.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	assert.Equal(t, "new@example.com", updated.Email)
	assert.Equal(t, memberCreator2, updated.LastModifierId)

	got, err := svc.GetMemberByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "New", got.Name)
	assert.Equal(t, "new@example.com", got.Email)
}

func TestMemberService_UpdateMember_NotFound(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMemberService(db)

	fakeID := uuid.MustParse("00000000-0000-0000-0000-00000000dead")
	m, err := svc.UpdateMember(fakeID, "X", "x@example.com", memberCreator1)
	assert.Error(t, err)
	assert.Nil(t, m)
	assert.Equal(t, "會員不存在", err.Error())
}

func TestMemberService_DeleteMember_Success(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMemberService(db)

	created, err := svc.CreateMember("Del", "del@example.com", "pw", memberCreator1)
	require.NoError(t, err)

	err = svc.DeleteMember(created.ID, memberCreator2)
	assert.NoError(t, err)

	got, err := svc.GetMemberByID(created.ID)
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, "會員不存在", err.Error())
}

func TestMemberService_DeleteMember_NotFoundOrAlreadyDeleted(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMemberService(db)

	fakeID := uuid.MustParse("00000000-0000-0000-0000-00000000cafe")
	err := svc.DeleteMember(fakeID, memberCreator1)
	assert.Error(t, err)
	assert.Equal(t, "會員不存在或已被刪除", err.Error())

	created, err := svc.CreateMember("Twice", "twice@example.com", "pw", memberCreator1)
	require.NoError(t, err)
	require.NoError(t, svc.DeleteMember(created.ID, memberCreator1))

	err = svc.DeleteMember(created.ID, memberCreator1)
	assert.Error(t, err)
	assert.Equal(t, "會員不存在或已被刪除", err.Error())
}

func TestMemberService_GetMemberByID_NotFound(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMemberService(db)

	fakeID := uuid.MustParse("00000000-0000-0000-0000-00000000babe")
	m, err := svc.GetMemberByID(fakeID)
	assert.Error(t, err)
	assert.Nil(t, m)
	assert.Equal(t, "會員不存在", err.Error())
}

func TestMemberService_GetMembers_EmptyAndList(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMemberService(db)

	list, err := svc.GetMembers(10)
	require.NoError(t, err)
	assert.Empty(t, list)

	insertMemberRow(t, db, "M1", "m1@example.com", memberCreator1)
	insertMemberRow(t, db, "M2", "m2@example.com", memberCreator1)

	list, err = svc.GetMembers(10)
	require.NoError(t, err)
	require.Len(t, list, 2)
}

func TestMemberService_GetMembers_RespectsLimit(t *testing.T) {
	db := testhelper.NewSQLiteTestDB(t)
	svc := NewMemberService(db)

	for i := 0; i < 5; i++ {
		email := uuid.New().String() + "@example.com"
		insertMemberRow(t, db, "U", email, memberCreator1)
	}

	list, err := svc.GetMembers(3)
	require.NoError(t, err)
	assert.Len(t, list, 3)
}
