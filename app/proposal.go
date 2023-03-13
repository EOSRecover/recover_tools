package app

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/msig"
	"github.com/eoscanada/eos-go/sudo"
	"github.com/eoscanada/eos-go/system"
	"github.com/pkg/errors"
	"recover_tool/conf"
	"recover_tool/logger"
	"time"
)

const (
	AccountLimit   = 100
	ProducersLimit = 21
)

func SendProposal() (err error) {
	
	// hacker accounts limit = 100
	// 200 actions (update owner and active)
	// Out of limit will be too long transaction
	if len(conf.APPConf().HackerAccounts) > AccountLimit {
		
		err = fmt.Errorf("out of account limit -> %d", AccountLimit)
		return
	}
	
	authTrx := BuildAuthTransaction()
	proposalAction, err := BuildProposal(authTrx)
	if err != nil {
		
		return
	}
	
	ctx := context.Background()
	txOpts := &eos.TxOptions{}
	if e := txOpts.FillFromChain(ctx, api); e != nil {
		
		logger.Instance().Error(" fill opts error -> ", e)
		return
	}
	
	tx := eos.NewTransaction([]*eos.Action{proposalAction}, txOpts)
	// txId, err := SendTx(tx, txOpts.ChainID)
	txId, err := SendTxByGM(tx, txOpts.ChainID)
	if err != nil {
		
		logger.Instance().Error("send Tx error -> ", err)
		return
	}
	
	logger.Instance().Info("send proposal succeed -> ", txId)
	
	return
}

func SendTx(tx *eos.Transaction, chainId eos.Checksum256) (txId string, err error) {
	
	ctx := context.Background()
	
	_, packedTx, err := api.SignTransaction(ctx, tx, chainId, eos.CompressionNone)
	if err != nil {
		
		logger.Instance().Error(" sign tx error -> ", err)
		return
	}
	
	response, err := api.PushTransaction(ctx, packedTx)
	if err != nil {
		
		logger.Instance().Error(" send tx error -> ", err)
		return
	}
	
	txId = hex.EncodeToString(response.Processed.ID)
	return
}

// SendTxByGM send tx by grey-mass resource api
func SendTxByGM(tx *eos.Transaction, chainId eos.Checksum256) (txId string, err error) {
	
	ctx := context.Background()
	
	client := &APIClient{
		Host: conf.APPConf().GMEndPoint,
	}
	
	freeTx, signers, err := client.RequestTxByGM(tx.Actions[0].Authorization[0], tx)
	if err != nil {
		
		err = errors.Wrap(err, "request grey-mass tx error")
		return
	}
	
	// read public keys
	pubKeys, err := api.Signer.AvailableKeys(ctx)
	if err != nil {
		
		err = errors.Wrap(err, "get pub keys error")
		return
	}
	
	// sign tx again
	tempTx, err := api.Signer.Sign(ctx, freeTx, chainId, pubKeys...)
	if err != nil {
		
		err = errors.Wrap(err, "sign tx error")
		return
	}
	
	// append signatures
	tempTx.Signatures = append(signers, tempTx.Signatures...)
	
	// repack
	newFreePackedTx, err := tempTx.Pack(eos.CompressionNone)
	if err != nil {
		
		err = errors.Wrap(err, "压缩交易出错")
		return
	}
	
	response, err := api.SendTransaction(ctx, newFreePackedTx)
	if err != nil {
		
		// logger.Instance().Error(" send tx error -> ", err)
		err = errors.Wrap(err, "send tx error")
		return
	}
	
	txId = hex.EncodeToString(response.Processed.ID)
	return
}

// BuildAuthTransaction build update auth actions
func BuildAuthTransaction() (transaction *eos.Transaction) {
	
	newPermission := NewAccountPermission()
	var actions []*eos.Action
	
	for _, account := range conf.APPConf().HackerAccounts {
		
		activeAction := system.NewUpdateAuth(
			eos.AccountName(account),
			"active",
			"owner",
			eos.Authority{
				Threshold: 1,
				Keys:      nil,
				Accounts:  newPermission,
				Waits:     nil,
			},
			"active",
		)
		
		ownerAction := system.NewUpdateAuth(
			eos.AccountName(account),
			"owner",
			"",
			eos.Authority{
				Threshold: 1,
				Keys:      nil,
				Accounts:  newPermission,
				Waits:     nil,
			},
			"owner",
		)
		
		actions = append(actions, activeAction)
		actions = append(actions, ownerAction)
	}
	
	transaction = eos.NewTransaction(actions, &eos.TxOptions{})
	transaction.SetExpiration(720 * time.Hour)
	
	return
}

func BuildProposal(tx *eos.Transaction) (proposalAction *eos.Action, err error) {
	
	if tx == nil {
		
		err = fmt.Errorf("tx is nil")
		return
	}
	
	bpsPermissions, err := BuildBPsPermission()
	if err != nil {
		
		return
	}
	
	// wrap
	wrapPermission, err := eos.NewPermissionLevel("eosio.wrap@active")
	if err != nil {
		
		return
	}
	execAction := sudo.NewExec("eosio", *tx)
	execAction.Authorization = append(execAction.Authorization, wrapPermission)
	
	// build proposal tx
	var actions []*eos.Action
	actions = append(actions, execAction)
	
	proposalTx := eos.NewTransaction(actions, &eos.TxOptions{})
	proposalTx.SetExpiration(time.Hour * 720)
	
	// build proposal
	proposalAction = msig.NewPropose(
		eos.AccountName(conf.APPConf().SendAccount),
		"freeze",
		bpsPermissions,
		proposalTx,
	)
	return
}

func BuildBPsPermission() (permissions []eos.PermissionLevel, err error) {
	
	bps, err := GetBPs()
	if err != nil {
		
		return
	}
	
	for _, bp := range bps {
		
		var permission eos.PermissionLevel
		permission, err = eos.NewPermissionLevel(bp)
		
		if err != nil {
			
			return
		}
		
		permissions = append(permissions, permission)
	}
	
	return
}

func NewAccountPermission() (accountPermission []eos.PermissionLevelWeight) {
	
	p, _ := eos.NewPermissionLevel("eosio")
	
	accountPermission = append(accountPermission, eos.PermissionLevelWeight{
		Permission: p,
		Weight:     1,
	})
	
	return
}

func GetBPs() (bps []string, err error) {
	
	resp, err := api.GetProducers(context.TODO())
	if err != nil {
		
		return
	}
	
	for index, p := range resp.Producers {
		
		if index >= ProducersLimit {
			
			break
		}
		
		bps = append(bps, p.Owner.String())
	}
	
	return
}
