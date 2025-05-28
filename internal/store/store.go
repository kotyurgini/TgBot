package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"tgbot/common"
	"time"
)

type Store struct {
	driver     Driver
	appSetting *AppSetting
	userCache  sync.Map // [int64] *UserShell
	aiModels   sync.Map // [int32] *AiModel
	tariffs    sync.Map // [int32] *TariffShell
}

const (
	DefaultAIChatModelID    int32 = 1
	DefaultAIImageModelID   int32 = 2
	DefaultTariffID         int32 = 1
	TgCheckNewDialogTimeout       = 3600 * time.Second
)

var (
	ErrIncorrectTariff  = errors.New("incorrect tariff")
	ErrIncorrectAIModel = errors.New("incorrect AI model")
)

func New(driver Driver) (*Store, error) {
	s := &Store{
		driver: driver,
	}

	err := s.LoadDefaultValues()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) LoadDefaultValues() error {
	var err error

	// Загрузка настроек приложения
	s.appSetting, err = s.driver.AppSettingGet(context.TODO())
	if err != nil {
		return fmt.Errorf("failed create Store: %w", err)
	}

	// Загрузка моделей
	aiModels, err := s.driver.AiModelList(context.TODO())
	if err != nil {
		return fmt.Errorf("failed create Store: %w", err)
	}
	for _, model := range aiModels {
		s.aiModels.Store(model.ID, model)
	}
	if _, ok := s.aiModels.Load(DefaultAIChatModelID); !ok {
		return fmt.Errorf("failed create Store: %w: %w", errors.New("default AI chat model not found"), err)
	}
	if _, ok := s.aiModels.Load(DefaultAIImageModelID); !ok {
		return fmt.Errorf("failed create Store: %w: %w", errors.New("default AI image model not found"), err)
	}

	// Загрузка тарифов
	trfs, err := s.driver.TariffList(context.TODO(), &TariffFilter{})
	if err != nil {
		return fmt.Errorf("failed create Store: %w: %w", errors.New("tariffs not found"), err)
	}
	tariffLimits, err := s.driver.TariffLimitList(context.TODO(), &TariffLimitFilter{})
	if err != nil {
		return fmt.Errorf("failed create Store: %w: %w", errors.New("tariff limits not found"), err)
	}

	for _, trf := range trfs {
		tariffShell := &TariffShell{Tariff: trf}
		for _, limit := range tariffLimits {
			if trf.ID == limit.TariffID {
				tariffShell.Limits = append(tariffShell.Limits, limit)
			}
		}
		s.tariffs.Store(trf.ID, tariffShell)
	}
	if _, ok := s.tariffs.Load(DefaultTariffID); !ok {
		return fmt.Errorf("failed create Store: %w", errors.New("default tariff not found"))
	}
	return nil
}

func (s *Store) GetDriver() Driver {
	return s.driver
}

func (s *Store) Close() error {
	return s.driver.Close()
}

func (s *Store) ValidateUser(ctx context.Context, userID int64, locale string) (*UserShell, error) {
	if cache, ok := s.userCache.Load(userID); ok {
		user, ok := cache.(*UserShell)
		if ok {
			user.Locale = locale
			err := s.CheckUserLimits(ctx, user)
			if err != nil {
				return nil, err
			}
			return user, nil
		}
	}

	var user *User

	users, err := s.driver.UserList(ctx, &UserFilter{ID: &userID})
	if err != nil {
		return nil, err
	}
	if len(users) > 0 {
		user = users[0]
	} else {
		user, err = s.driver.UserCreate(ctx, &User{ID: userID, ChatModelID: DefaultAIChatModelID, ImageModelID: DefaultAIImageModelID, TariffID: DefaultTariffID})
		if err != nil {
			return nil, err
		}
	}

	us := &UserShell{
		User:     user,
		Dialog:   nil,
		Context:  nil,
		LastText: "",
		Locale:   locale,
	}

	// Проверка активного диалога
	activeDialogList, err := s.driver.ActiveDialogList(ctx, &ActiveDialogFilter{UserID: &userID})
	if err != nil {
		return nil, err
	}
	// Получение активного диалога
	if len(activeDialogList) > 0 {
		_, dialogs, err := s.driver.DialogList(ctx, &DialogFilter{ID: &activeDialogList[0].DialogID})
		if err != nil {
			return nil, err
		}
		if len(dialogs) > 0 {
			us.Dialog = dialogs[0]
		}
	}

	// Получение сообщений диалога
	if us.Dialog != nil {
		msgs, err := s.driver.ChatMessageList(ctx, &ChatMessageFilter{DialogID: &us.Dialog.ID})
		if err != nil {
			return nil, err
		}
		if len(msgs) > 0 {
			us.Context = msgs
		}
	}

	// Получение информации об использований
	userUsageList, err := s.driver.UserUsageList(ctx, &UserUsageFilter{UserID: &userID})
	if err != nil {
		return nil, err
	}
	for _, usage := range userUsageList {
		us.Usage.Store(usage.AIModelID, usage)
	}

	us.ID = user.ID
	s.userCache.Store(user.ID, us)

	err = s.CheckUserLimits(ctx, us)
	if err != nil {
		return nil, err
	}

	return us, nil
}

func (s *Store) UserDialogs(ctx context.Context, filter *DialogFilter) (int, []*Dialog, error) {
	count, dialogs, err := s.driver.DialogList(ctx, filter)
	if err != nil {
		return 0, nil, err
	}
	return count, dialogs, nil
}

func (s *Store) AddDialog(ctx context.Context, user *UserShell, dialog *Dialog) (*Dialog, error) {
	dialog, err := s.driver.DialogCreate(ctx, dialog)
	if err != nil {
		return nil, err
	}

	_, err = s.driver.ActiveDialogUpsert(ctx, &ActiveDialog{UserID: user.ID, DialogID: dialog.ID})
	if err != nil {
		return nil, err
	}

	user.Dialog = dialog

	return dialog, nil
}

func (s *Store) DeleteDialog(ctx context.Context, user *UserShell, filter *DialogFilter) error {
	err := s.driver.DialogDelete(ctx, filter)
	if err != nil {
		return err
	}

	// TODO : Если диалог удален без id в filter, то как определить является ли он активным для пользователя?
	if user.Dialog != nil && ((filter.ID != nil && user.Dialog.ID == *filter.ID) || filter.UserID != nil) {
		user.Dialog = nil
		user.Context = nil
	}

	return nil
}

func (s *Store) DialogInfo(ctx context.Context, dialogID int64) (*DialogInfo, error) {
	_, dialogs, err := s.driver.DialogList(ctx, &DialogFilter{ID: &dialogID})
	if err != nil {
		return nil, err
	}
	if len(dialogs) == 0 {
		return nil, errors.New("dialog not found")
	}

	msgs, err := s.driver.ChatMessageList(ctx, &ChatMessageFilter{DialogID: &dialogID})
	if err != nil {
		return nil, err
	}

	return &DialogInfo{
		Dialog:  dialogs[0],
		Context: msgs,
	}, nil
}

func (s *Store) UpdateActiveDialog(ctx context.Context, user *UserShell, dialogInfo *DialogInfo) error {
	_, err := s.driver.ActiveDialogUpsert(ctx, &ActiveDialog{UserID: user.ID, DialogID: dialogInfo.Dialog.ID})
	if err != nil {
		return err
	}

	user.Dialog = dialogInfo.Dialog
	user.Context = dialogInfo.Context
	return nil
}

func (s *Store) ResetActiveDialog(ctx context.Context, user *UserShell) error {
	err := s.driver.ActiveDialogDelete(ctx, &ActiveDialogFilter{UserID: &user.ID})
	if err != nil {
		return err
	}

	user.Dialog = nil
	user.Context = nil
	return nil
}

func (s *Store) AddNewMessage(ctx context.Context, user *UserShell, msg *ChatMessage) (*ChatMessage, error) {
	msg, err := s.driver.ChatMessageCreate(ctx, msg)
	if err != nil {
		return nil, err
	}

	user.Context = append(user.Context, msg)
	return msg, nil
}

func (s *Store) Tariffs() []*TariffShell {
	var result []*TariffShell
	s.tariffs.Range(func(_, value any) bool {
		if ts, ok := value.(*TariffShell); ok {
			result = append(result, ts)
		}
		return true
	})
	return result
}

func (s *Store) TariffByID(id int32) (*TariffShell, bool) {
	trf, ok := s.tariffs.Load(id)
	return trf.(*TariffShell), ok
}

func (s *Store) AIModelByID(id int32) (*AiModel, bool) {
	model, ok := s.aiModels.Load(id)
	return model.(*AiModel), ok
}

func (s *Store) UserUsageToString(tariffLimits []*TariffLimit, userUsage []*UserUsage) string {
	if len(tariffLimits) == 0 {
		return ""
	}
	var result strings.Builder
	for _, limit := range tariffLimits {
		model, _ := s.AIModelByID(limit.AIModelID)
		var usageCount int32 = 0
		for _, usage := range userUsage {
			if usage.AIModelID == limit.AIModelID {
				usageCount = usage.Count
			}
		}
		result.WriteString(fmt.Sprintf("%s: (%d/%d)\n", model.Title, usageCount, limit.Count))
	}
	return result.String()
}

func (s *Store) CheckUserUsage(user *UserShell, modelID int32) (bool, error) {
	tariff, ok := s.TariffByID(user.User.TariffID)
	if !ok {
		return false, ErrIncorrectTariff
	}

	var limit *TariffLimit
	for _, l := range tariff.Limits {
		if l.AIModelID == modelID {
			limit = l
			break
		}
	}
	if limit == nil {
		return false, fmt.Errorf("CheckUserUsage() limit not found by id %d", modelID)
	}

	fUsage, ok := user.Usage.Load(modelID)
	if !ok {
		return false, fmt.Errorf("CheckUserUsage() usage not found for model id %d", modelID)
	}
	usage := fUsage.(*UserUsage)

	if usage.Count >= limit.Count {
		return false, nil
	}

	return true, nil
}

func (s *Store) UpdateUserUsage(ctx context.Context, user *UserShell, modelID int32) error {
	fUsage, ok := user.Usage.Load(modelID)
	if !ok {
		return fmt.Errorf("UpdateUserUsage() usage not found for model id %d", modelID)
	}
	usage := fUsage.(*UserUsage)
	usage.Count++
	usage.LastActivity = time.Now().UTC()
	_, err := s.driver.UserUsageUpdate(ctx, usage)
	if err != nil {
		usage.Count--
		return err
	}

	return nil
}

func (s *Store) CheckUserLastActivity(user *UserShell) bool {
	result := false
	user.Usage.Range(func(key, value any) bool {
		if uu, ok := value.(*UserUsage); ok {
			if time.Since(uu.LastActivity) <= TgCheckNewDialogTimeout {
				result = true
				return false
			}
		}
		return true
	})

	return !result
}

func (s *Store) CheckUserLimits(ctx context.Context, user *UserShell) error {
	now := common.TimeNowUTCDay()
	if user.User.LastLimitReset.Equal(now) {
		return nil
	}

	var usage []*UserUsage
	user.Usage.Range(func(_, v any) bool {
		if val, ok := v.(*UserUsage); ok {
			usage = append(usage, val)
		}
		return true
	})

	for _, u := range usage {
		prevCount := u.Count
		u.Count = 0
		_, err := s.driver.UserUsageUpdate(ctx, u)
		if err != nil {
			u.Count = prevCount
			return err
		}
	}
	user.User.LastLimitReset = now
	_, err := s.driver.UserUpdate(ctx, user.User)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) SetSelfBlockUser(us *UserShell, block bool) {
	if us.User.SelfBlock == block {
		return
	}
	us.User.SelfBlock = block
	_, _ = s.driver.UserUpdate(context.TODO(), us.User)
}

func (s *Store) BlockUser(ctx context.Context, us *UserShell, reason string, adminID int64) error {
	if us.User.ID == adminID {
		return fmt.Errorf("BlockUser(): %w", errors.New("admin can't be blocked"))
	}

	us.User.Blocked = true
	us.User.BlockReason = reason
	_, err := s.driver.UserUpdate(ctx, us.User)
	if err != nil {
		return fmt.Errorf("BlockUser(): %w", err)
	}
	return nil
}

func (s *Store) UnblockUser(ctx context.Context, us *UserShell) error {
	us.User.Blocked = false
	us.User.BlockReason = ""
	_, err := s.driver.UserUpdate(ctx, us.User)
	if err != nil {
		return fmt.Errorf("UnblockUser(): %w", err)
	}
	return nil
}

// In some cases might return UserShell without additional information, but always with user db entity
func (s *Store) GetUserShellByID(ctx context.Context, userID int64) (*UserShell, error) {
	if cache, ok := s.userCache.Load(userID); ok {
		if user, ok := cache.(*UserShell); ok {
			return user, nil
		}
	}

	users, err := s.driver.UserList(ctx, &UserFilter{ID: &userID})
	if err != nil {
		return nil, err
	}
	if len(users) > 0 {
		user := users[0]
		us := &UserShell{
			User: user,
		}
		return us, nil
	} else {
		return nil, fmt.Errorf("GetUserShellByID(): %w", errors.New("user not found"))
	}
}

func (s *Store) ToggleUserSkipDialog(ctx context.Context, us *UserShell) error {
	us.User.SkipNewDialogMessage = !us.User.SkipNewDialogMessage
	_, err := s.driver.UserUpdate(ctx, us.User)
	if err != nil {
		return fmt.Errorf("toggleUserSkipDialog(): %w", err)
	}
	return nil
}
