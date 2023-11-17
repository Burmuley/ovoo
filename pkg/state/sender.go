package state

func (s *State) Senders() ([]Sender, error) {
	var senderEmails []Sender
	err := s.db.Model(&Sender{}).Find(&senderEmails).Error
	if err != nil {
		return nil, err
	}
	return senderEmails, err
}

func (s *State) CreateSender(e Sender) error {
	err := s.db.Model(&Sender{}).Create(&e).Error
	return err
}

func (s *State) DeleteSender(se Sender) error {
	return s.db.Delete(&se).Error
}

func (s *State) GetSenderByEmail(email string) (Sender, bool) {
	se := &Sender{}
	err := s.db.Model(&Sender{}).Where("email = ?", email).First(&se).Error
	if err != nil {
		return Sender{}, false
	}

	return *se, true
}

func (s *State) GetSenderById(id string) (Sender, bool) {
	se := &Sender{}
	err := s.db.Model(&Sender{}).Where("id = ?", id).First(&se).Error
	if err != nil {
		return Sender{}, false
	}

	return *se, true
}

func (s *State) GetSenderByReplyAliasEmail(email string) (Sender, bool) {
	panic("implement me!")
}
