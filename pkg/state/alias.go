package state

func (s *State) Aliases() ([]Alias, error) {
	var aes []Alias
	err := s.db.Model(&Alias{}).Preload("ProtectedAddress").Find(&aes).Error
	if err != nil {
		return nil, err
	}
	return aes, err
}

func (s *State) CreateAlias(ae Alias) (Alias, error) {
	if err := s.db.Model(&Alias{}).Create(&ae).Error; err != nil {
		return Alias{}, err
	}

	alias, ok := s.GetAliasByEmail(ae.Email)
	if !ok {
		return Alias{}, ErrRecordNotFound
	}

	return alias, nil
}

func (s *State) DeleteAliasByEmail(e string) error {
	return s.db.Model(&Alias{}).Where("email = ?", e).Unscoped().Delete(&Alias{Email: e}).Error
}

func (s *State) DeleteAliasById(id string) error {
	return s.db.Model(&Alias{}).Where("id = ?", id).Unscoped().Delete(&Alias{Model: Model{ID: id}}).Error
}

func (s *State) GetAliasByEmail(email string) (Alias, bool) {
	ae := &Alias{}
	err := s.db.Model(&Alias{}).Preload("ProtectedAddress").Where("email = ?", email).First(ae).Error
	if err != nil {
		return Alias{}, false
	}

	return *ae, true
}

func (s *State) GetAliasById(id string) (Alias, bool) {
	ae := &Alias{}
	err := s.db.Model(&Alias{}).Preload("ProtectedAddress").Where("id = ?", id).First(&ae).Error
	if err != nil {
		return Alias{}, false
	}

	return *ae, true
}
