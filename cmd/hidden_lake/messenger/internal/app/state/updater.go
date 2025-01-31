package state

import "fmt"

func (p *sState) stateUpdater(
	clientUpdater func(storageValue *SStorageState) error,
	middleWare func(storageValue *SStorageState),
) error {
	p.fMutex.Lock()
	defer p.fMutex.Unlock()

	if !p.IsActive() {
		return fmt.Errorf("state does not exist")
	}

	oldStorageValue, err := p.getStorageState(p.fHashLP)
	if err != nil {
		return err
	}

	newStorageValue := copyStorageState(oldStorageValue)
	middleWare(newStorageValue)

	if err := clientUpdater(newStorageValue); err == nil {
		return p.setStorageState(newStorageValue)
	}

	return p.setStorageState(oldStorageValue)
}

func copyStorageState(pStorageValue *SStorageState) *SStorageState {
	copyStorageValue := &SStorageState{
		FPrivKey:     pStorageValue.FPrivKey,
		FFriends:     make(map[string]string, len(pStorageValue.FFriends)),
		FConnections: make([]string, 0, len(pStorageValue.FConnections)),
	}

	for aliasName, pubKey := range pStorageValue.FFriends {
		copyStorageValue.FFriends[aliasName] = pubKey
	}

	copyStorageValue.FConnections = append(
		copyStorageValue.FConnections,
		pStorageValue.FConnections...,
	)

	return copyStorageValue
}
