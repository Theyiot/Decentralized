package gossiper

import (
	"sync"
)

type TransactionsSet struct {
	transactions	[]*TxPublish
	lock			sync.RWMutex
}

func (set *TransactionsSet) contains(newTransaction *TxPublish) bool {
	set.lock.RLock()
	defer set.lock.RUnlock()
	for _, transaction := range set.transactions {
		if transaction.File.Name == newTransaction.File.Name {
			return true
		}
	}
	return false
}

func (set *TransactionsSet) Add(newTransaction *TxPublish) {
	set.lock.Lock()
	defer set.lock.Unlock()
	if !set.contains(newTransaction) {
		set.transactions = append(set.transactions, newTransaction)
	}
}

func (set *TransactionsSet) flushFromBlock(block Block) {
	newTransactions := make([]*TxPublish, 0)
	set.lock.Lock()
	defer set.lock.Unlock()
	for _, transaction := range set.transactions {
		transactionDone := false
		for _, blockTransaction := range block.Transactions {
			if transaction.File.Name == blockTransaction.File.Name {
				transactionDone = true
			}
		}
		if !transactionDone {
			newTransactions = append(newTransactions, transaction)
		}
	}
}

func (set *TransactionsSet) getSetCopy() []*TxPublish {
	var transactionsCopy []*TxPublish
	set.lock.RLock()
	defer set.lock.RUnlock()
	copy(transactionsCopy, set.transactions)
	return transactionsCopy
}

func createTransactionsSet() *TransactionsSet {
	return &TransactionsSet{ transactions:make([]*TxPublish, 0) }
}