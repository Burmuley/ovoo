package state

func (s *State) ReplyAliases() ([]ReplyAlias, error) {
	var raes []ReplyAlias
	err := s.db.Model(&ReplyAlias{}).Preload("Alias").
		Preload("Sender").Find(&raes).Error
	if err != nil {
		return nil, err
	}
	return raes, err
}

func (s *State) CreateReplyAlias(rae ReplyAlias) error {
	return s.db.Model(&ReplyAlias{}).Create(&rae).Error
}

func (s *State) DeleteReplyAlias(rae ReplyAlias) error {
	return s.db.Delete(&rae).Error
}

func (s *State) GetReplyAliasByEmail(email string) (ReplyAlias, bool) {
	rae := &ReplyAlias{}
	err := s.db.Model(&ReplyAlias{}).Where("email = ?", email).First(&rae).Error
	if err != nil {
		return ReplyAlias{}, false
	}

	return *rae, true
}

func (s *State) GetReplyAliasById(id string) (ReplyAlias, bool) {
	rae := &ReplyAlias{}
	err := s.db.Model(&ReplyAlias{}).Where("id = ?", id).First(&rae).Error
	if err != nil {
		return ReplyAlias{}, false
	}

	return *rae, true
}
