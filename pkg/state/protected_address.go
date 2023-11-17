package state

func (s *State) ProtectedAddresses() ([]ProtectedAddress, error) {
	var pas []ProtectedAddress
	err := s.db.Model(&ProtectedAddress{}).Find(&pas).Error
	if err != nil {
		return nil, err
	}
	return pas, nil
}

func (s *State) CreateProtectedAddress(pe ProtectedAddress) (ProtectedAddress, error) {
	if err := s.db.Model(&ProtectedAddress{}).Create(&pe).Error; err != nil {
		return ProtectedAddress{}, err
	}

	pa, ok := s.GetProtectedAddressByEmail(pe.Email)
	if !ok {
		return ProtectedAddress{}, ErrRecordNotFound
	}

	return pa, nil
}

func (s *State) DeleteProtectedAddressByEmail(e string) error {
	return s.db.Model(&ProtectedAddress{}).Where("email = ?", e).Unscoped().
		Delete(&ProtectedAddress{Email: e}).Error
}

func (s *State) DeleteProtectedAddressById(id string) error {
	return s.db.Model(&ProtectedAddress{}).Where("id = ?", id).Unscoped().
		Delete(&ProtectedAddress{Model: Model{ID: id}}).Error
}

func (s *State) GetProtectedAddressByEmail(email string) (ProtectedAddress, bool) {
	pa := &ProtectedAddress{}
	err := s.db.Model(&ProtectedAddress{}).Where("email = ?", email).First(&pa).Error
	if err != nil {
		return ProtectedAddress{}, false
	}

	return *pa, true
}

func (s *State) GetProtectedAddressById(id string) (ProtectedAddress, bool) {
	pa := &ProtectedAddress{}
	err := s.db.Model(&ProtectedAddress{}).Where("id = ?", id).First(&pa).Error
	if err != nil {
		return ProtectedAddress{}, false
	}

	return *pa, true
}
