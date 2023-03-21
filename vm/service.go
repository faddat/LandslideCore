package vm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmmath "github.com/tendermint/tendermint/libs/math"
	tmquery "github.com/tendermint/tendermint/libs/pubsub/query"
	mempl "github.com/tendermint/tendermint/mempool"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/rpc/core"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

type (
	ABCIInfoArgs struct {
	}

	ABCIQueryArgs struct {
		Path string           `json:"path"`
		Data tmbytes.HexBytes `json:"data"`
	}

	ABCIQueryOptions struct {
		Height int64 `json:"height"`
		Prove  bool  `json:"prove"`
	}

	ABCIQueryWithOptionsArgs struct {
		Path string           `json:"path"`
		Data tmbytes.HexBytes `json:"data"`
		Opts ABCIQueryOptions `json:"opts"`
	}

	// BroadcastTxArgs is the arguments to functions BroadcastTxCommit, BroadcastTxAsync, BroadcastTxSync
	BroadcastTxArgs struct {
		Tx types.Tx `json:"tx"`
	}

	BlockHeightArgs struct {
		Height *int64 `json:"height"`
	}

	BlockHashArgs struct {
		Hash []byte `json:"hash"`
	}

	ABCIInfoArgs struct {
	}

	ABCIQueryArgs struct {
		Path string           `json:"path"`
		Data tmbytes.HexBytes `json:"data"`
	}

	ABCIQueryOptions struct {
		Height int64 `json:"height"`
		Prove  bool  `json:"prove"`
	}

	ABCIQueryWithOptionsArgs struct {
		Path string           `json:"path"`
		Data tmbytes.HexBytes `json:"data"`
		Opts ABCIQueryOptions `json:"opts"`
  }

	BlockchainInfoArgs struct {
		MinHeight int64 `json:"minHeight"`
		MaxHeight int64 `json:"maxHeight"`
	}

	GenesisChunkedArgs struct {
		Chunk uint `json:"chunk"`
	}

	CommitArgs struct {
		Height *int64 `json:"height"`
	}

	ValidatorsArgs struct {
		Height  *int64 `json:"height"`
		Page    *int   `json:"page"`
		PerPage *int   `json:"perPage"`
	}

	TxArgs struct {
		Hash  []byte `json:"hash"`
		Prove bool   `json:"prove"`
	}

	TxSearchArgs struct {
		Query   string `json:"query"`
		Prove   bool   `json:"prove"`
		Page    *int   `json:"page"`
		PerPage *int   `json:"perPage"`
		OrderBy string `json:"orderBy"`
	}

	BlockSearchArgs struct {
		Query   string `json:"query"`
		Page    *int   `json:"page"`
		PerPage *int   `json:"perPage"`
		OrderBy string `json:"orderBy"`
	}

	StatusArgs struct{}

	NetInfoArgs struct{}

	DumpConsensusStateArgs struct{}

	ConsensusStateArgs struct{}

	ConsensusParamsArgs struct {
		Height *int64 `json:"height"`
	}

	HealthArgs struct{}

	UnconfirmedTxsArgs struct {
		Limit *int `json:"limit"`
	}

	NumUnconfirmedTxsArgs struct{}

	CheckTxArgs struct {
		Tx []byte `json:"tx"`
	}

	// Service is the API service for this VM
	Service interface {
		// Reading from abci app
		ABCIInfo(_ *http.Request, args *ABCIInfoArgs, reply *ctypes.ResultABCIInfo) error
		ABCIQuery(_ *http.Request, args *ABCIQueryArgs, reply *ctypes.ResultABCIQuery) error
		ABCIQueryWithOptions(_ *http.Request, args *ABCIQueryWithOptionsArgs, reply *ctypes.ResultABCIQuery) error

		// Writing to abci app
		BroadcastTxCommit(_ *http.Request, args *BroadcastTxArgs, reply *ctypes.ResultBroadcastTxCommit) error
		BroadcastTxAsync(_ *http.Request, args *BroadcastTxArgs, reply *ctypes.ResultBroadcastTx) error
		BroadcastTxSync(_ *http.Request, args *BroadcastTxArgs, reply *ctypes.ResultBroadcastTx) error

		Block(_ *http.Request, args *BlockHeightArgs, reply *ctypes.ResultBlock) error
		BlockByHash(_ *http.Request, args *BlockHashArgs, reply *ctypes.ResultBlock) error
		BlockResults(_ *http.Request, args *BlockHeightArgs, reply *ctypes.ResultBlockResults) error
		Commit(_ *http.Request, args *CommitArgs, reply *ctypes.ResultCommit) error
		Validators(_ *http.Request, args *ValidatorsArgs, reply *ctypes.ResultValidators) error
		Tx(_ *http.Request, args *TxArgs, reply *ctypes.ResultTx) error
		TxSearch(_ *http.Request, args *TxSearchArgs, reply *ctypes.ResultTxSearch) error
		BlockSearch(_ *http.Request, args *BlockSearchArgs, reply *ctypes.ResultBlockSearch) error

		BlockchainInfo(_ *http.Request, args *BlockchainInfoArgs, reply *ctypes.ResultBlockchainInfo) error

		Genesis(_ *http.Request, _ *struct{}, reply *ctypes.ResultGenesis) error
		GenesisChunked(_ *http.Request, args *GenesisChunkedArgs, reply *ctypes.ResultGenesisChunk) error

		Status(_ *http.Request, args *StatusArgs, reply *ctypes.ResultStatus) error

		NetInfo(_ *http.Request, args *NetInfoArgs, reply *ctypes.ResultNetInfo) error
		DumpConsensusState(_ *http.Request, args *DumpConsensusStateArgs, reply *ctypes.ResultDumpConsensusState) error
		ConsensusState(_ *http.Request, args *ConsensusStateArgs, reply *ctypes.ResultConsensusState) error
		ConsensusParams(_ *http.Request, args *ConsensusParamsArgs, reply *ctypes.ResultConsensusParams) error
		Health(_ *http.Request, args *HealthArgs, reply *ctypes.ResultHealth) error

		UnconfirmedTxs(_ *http.Request, args *UnconfirmedTxsArgs, reply *ctypes.ResultUnconfirmedTxs) error
		NumUnconfirmedTxs(_ *http.Request, args *NumUnconfirmedTxsArgs, reply *ctypes.ResultUnconfirmedTxs) error
		CheckTx(_ *http.Request, args *CheckTxArgs, reply *ctypes.ResultCheckTx) error
	}

	LocalService struct {
		vm *VM
	}
)

var (
	DefaultABCIQueryOptions = ABCIQueryOptions{Height: 0, Prove: false}
)

func NewService(vm *VM) Service {
	return &LocalService{vm}
}

func (s *LocalService) ABCIInfo(_ *http.Request, args *ABCIInfoArgs, reply *ctypes.ResultABCIInfo) error {
	resInfo, err := s.vm.proxyApp.Query().InfoSync(proxy.RequestInfo)
	if err != nil {
		return err
	}
	reply.Response = *resInfo
	return nil
}

func (s *LocalService) ABCIQuery(req *http.Request, args *ABCIQueryArgs, reply *ctypes.ResultABCIQuery) error {
	return s.ABCIQueryWithOptions(req, &ABCIQueryWithOptionsArgs{args.Path, args.Data, DefaultABCIQueryOptions}, reply)
}

func (s *LocalService) ABCIQueryWithOptions(
	_ *http.Request,
	args *ABCIQueryWithOptionsArgs,
	reply *ctypes.ResultABCIQuery,
) error {
	resQuery, err := s.vm.proxyApp.Query().QuerySync(abci.RequestQuery{
		Path:   args.Path,
		Data:   args.Data,
		Height: args.Opts.Height,
		Prove:  args.Opts.Prove,
	})
	if err != nil {
		return err
	}
	reply.Response = *resQuery
	return nil
}

func (s *LocalService) BroadcastTxCommit(
	_ *http.Request,
	args *BroadcastTxArgs,
	reply *ctypes.ResultBroadcastTxCommit,
) error {
	subscriber := ""

	// Subscribe to tx being committed in block.
	subCtx, cancel := context.WithTimeout(context.Background(), core.SubscribeTimeout)
	defer cancel()

	q := types.EventQueryTxFor(args.Tx)
	deliverTxSub, err := s.vm.eventBus.Subscribe(subCtx, subscriber, q)
	if err != nil {
		err = fmt.Errorf("failed to subscribe to tx: %w", err)
		s.vm.tmLogger.Error("Error on broadcast_tx_commit", "err", err)
		return err
	}

	defer func() {
		if err := s.vm.eventBus.Unsubscribe(context.Background(), subscriber, q); err != nil {
			s.vm.tmLogger.Error("Error unsubscribing from eventBus", "err", err)
		}
	}()

	// Broadcast tx and wait for CheckTx result
	checkTxResCh := make(chan *abci.Response, 1)
	err = s.vm.mempool.CheckTx(args.Tx, func(res *abci.Response) {
		checkTxResCh <- res
	}, mempl.TxInfo{})
	if err != nil {
		s.vm.tmLogger.Error("Error on broadcastTxCommit", "err", err)
		return fmt.Errorf("error on broadcastTxCommit: %v", err)
	}
	checkTxResMsg := <-checkTxResCh
	checkTxRes := checkTxResMsg.GetCheckTx()
	if checkTxRes.Code != abci.CodeTypeOK {
		reply.CheckTx = *checkTxRes
		reply.DeliverTx = abci.ResponseDeliverTx{}
		reply.Hash = args.Tx.Hash()
		return nil
	}

	// Wait for the tx to be included in a block or timeout.
	select {
	case msg := <-deliverTxSub.Out(): // The tx was included in a block.
		deliverTxRes := msg.Data().(types.EventDataTx)
		reply.CheckTx = *checkTxRes
		reply.DeliverTx = deliverTxRes.Result
		reply.Hash = args.Tx.Hash()
		reply.Height = deliverTxRes.Height
		return nil
	case <-deliverTxSub.Cancelled():
		var reason string
		if deliverTxSub.Err() == nil {
			reason = "Tendermint exited"
		} else {
			reason = deliverTxSub.Err().Error()
		}
		err = fmt.Errorf("deliverTxSub was cancelled (reason: %s)", reason)
		s.vm.tmLogger.Error("Error on broadcastTxCommit", "err", err)
		return err
	// TODO: use config for timeout
	case <-time.After(10 * time.Second):
		err = errors.New("timed out waiting for tx to be included in a block")
		s.vm.tmLogger.Error("Error on broadcastTxCommit", "err", err)
		return err
	}
}

func (s *LocalService) BroadcastTxAsync(
	_ *http.Request,
	args *BroadcastTxArgs,
	reply *ctypes.ResultBroadcastTx,
) error {
	err := s.vm.mempool.CheckTx(args.Tx, nil, mempl.TxInfo{})
	if err != nil {
		return err
	}
	reply.Hash = args.Tx.Hash()
	return nil
}

func (s *LocalService) BroadcastTxSync(_ *http.Request, args *BroadcastTxArgs, reply *ctypes.ResultBroadcastTx) error {
	resCh := make(chan *abci.Response, 1)
	err := s.vm.mempool.CheckTx(args.Tx, func(res *abci.Response) {
		resCh <- res
	}, mempl.TxInfo{})
	if err != nil {
		return err
	}
	res := <-resCh
	r := res.GetCheckTx()

	reply.Code = r.Code
	reply.Data = r.Data
	reply.Log = r.Log
	reply.Codespace = r.Codespace
	reply.Hash = args.Tx.Hash()

	return nil
}

func (s *LocalService) Block(_ *http.Request, args *BlockHeightArgs, reply *ctypes.ResultBlock) error {
	height, err := getHeight(s.vm.blockStore, args.Height)
	if err != nil {
		return err
	}
	block := s.vm.blockStore.LoadBlock(height)
	blockMeta := s.vm.blockStore.LoadBlockMeta(height)

	if blockMeta != nil {
		reply.BlockID = blockMeta.BlockID
	}
	reply.Block = block
	return nil
}

func (s *LocalService) BlockByHash(_ *http.Request, args *BlockHashArgs, reply *ctypes.ResultBlock) error {
	block := s.vm.blockStore.LoadBlockByHash(args.Hash)
	if block == nil {
		reply.BlockID = types.BlockID{}
		reply.Block = nil
		return nil
	}
	blockMeta := s.vm.blockStore.LoadBlockMeta(block.Height)
	reply.BlockID = blockMeta.BlockID
	reply.Block = block
	return nil
}

func (s *LocalService) BlockResults(_ *http.Request, args *BlockHeightArgs, reply *ctypes.ResultBlockResults) error {
	height, err := getHeight(s.vm.blockStore, args.Height)
	if err != nil {
		return err
	}

	results, err := s.vm.stateStore.LoadABCIResponses(height)
	if err != nil {
		return err
	}

	reply.Height = height
	reply.TxsResults = results.DeliverTxs
	reply.BeginBlockEvents = results.BeginBlock.Events
	reply.EndBlockEvents = results.EndBlock.Events
	reply.ValidatorUpdates = results.EndBlock.ValidatorUpdates
	reply.ConsensusParamUpdates = results.EndBlock.ConsensusParamUpdates
	return nil
}

func (s *LocalService) Commit(_ *http.Request, args *CommitArgs, reply *ctypes.ResultCommit) error {
	height, err := getHeight(s.vm.blockStore, args.Height)
	if err != nil {
		return err
	}

	blockMeta := s.vm.blockStore.LoadBlockMeta(height)
	if blockMeta == nil {
		return nil
	}

	header := blockMeta.Header
	commit := s.vm.blockStore.LoadBlockCommit(height)
	res := ctypes.NewResultCommit(&header, commit, !(height == s.vm.blockStore.Height()))

	reply.SignedHeader = res.SignedHeader
	reply.CanonicalCommit = res.CanonicalCommit
	return nil
}

func (s *LocalService) Validators(_ *http.Request, args *ValidatorsArgs, reply *ctypes.ResultValidators) error {
	height, err := getHeight(s.vm.blockStore, args.Height)
	if err != nil {
		return err
	}

	validators, err := s.vm.stateStore.LoadValidators(height)
	if err != nil {
		return err
	}

	totalCount := len(validators.Validators)
	perPage := validatePerPage(args.PerPage)
	page, err := validatePage(args.Page, perPage, totalCount)
	if err != nil {
		return err
	}

	skipCount := validateSkipCount(page, perPage)

	reply.BlockHeight = height
	reply.Validators = validators.Validators[skipCount : skipCount+tmmath.MinInt(perPage, totalCount-skipCount)]
	reply.Count = len(reply.Validators)
	reply.Total = totalCount
	return nil
}

func (s *LocalService) Tx(_ *http.Request, args *TxArgs, reply *ctypes.ResultTx) error {
	r, err := s.vm.txIndexer.Get(args.Hash)
	if err != nil {
		return err
	}

	if r == nil {
		return fmt.Errorf("tx (%X) not found", args.Hash)
	}

	height := r.Height
	index := r.Index

	var proof types.TxProof
	if args.Prove {
		block := s.vm.blockStore.LoadBlock(height)
		proof = block.Data.Txs.Proof(int(index)) // XXX: overflow on 32-bit machines
	}

	reply.Hash = args.Hash
	reply.Height = height
	reply.Index = index
	reply.TxResult = r.Result
	reply.Tx = r.Tx
	reply.Proof = proof
	return nil
}

func (s *LocalService) TxSearch(req *http.Request, args *TxSearchArgs, reply *ctypes.ResultTxSearch) error {
	q, err := tmquery.New(args.Query)
	if err != nil {
		return err
	}

	results, err := s.vm.txIndexer.Search(req.Context(), q)
	if err != nil {
		return err
	}

	// sort results (must be done before pagination)
	switch args.OrderBy {
	case "desc":
		sort.Slice(results, func(i, j int) bool {
			if results[i].Height == results[j].Height {
				return results[i].Index > results[j].Index
			}
			return results[i].Height > results[j].Height
		})
	case "asc", "":
		sort.Slice(results, func(i, j int) bool {
			if results[i].Height == results[j].Height {
				return results[i].Index < results[j].Index
			}
			return results[i].Height < results[j].Height
		})
	default:
		return errors.New("expected order_by to be either `asc` or `desc` or empty")
	}

	// paginate results
	totalCount := len(results)
	perPage := validatePerPage(args.PerPage)

	page, err := validatePage(args.Page, perPage, totalCount)
	if err != nil {
		return err
	}

	skipCount := validateSkipCount(page, perPage)
	pageSize := tmmath.MinInt(perPage, totalCount-skipCount)

	apiResults := make([]*ctypes.ResultTx, 0, pageSize)
	for i := skipCount; i < skipCount+pageSize; i++ {
		r := results[i]

		var proof types.TxProof
		if args.Prove {
			block := s.vm.blockStore.LoadBlock(r.Height)
			proof = block.Data.Txs.Proof(int(r.Index)) // XXX: overflow on 32-bit machines
		}

		apiResults = append(apiResults, &ctypes.ResultTx{
			Hash:     types.Tx(r.Tx).Hash(),
			Height:   r.Height,
			Index:    r.Index,
			TxResult: r.Result,
			Tx:       r.Tx,
			Proof:    proof,
		})
	}

	reply.Txs = apiResults
	reply.TotalCount = totalCount
	return nil
}

func (s *LocalService) BlockSearch(req *http.Request, args *BlockSearchArgs, reply *ctypes.ResultBlockSearch) error {
	q, err := tmquery.New(args.Query)
	if err != nil {
		return err
	}

	results, err := s.vm.blockIndexer.Search(req.Context(), q)
	if err != nil {
		return err
	}

	// sort results (must be done before pagination)
	switch args.OrderBy {
	case "desc", "":
		sort.Slice(results, func(i, j int) bool { return results[i] > results[j] })

	case "asc":
		sort.Slice(results, func(i, j int) bool { return results[i] < results[j] })

	default:
		return errors.New("expected order_by to be either `asc` or `desc` or empty")
	}

	// paginate results
	totalCount := len(results)
	perPage := validatePerPage(args.PerPage)

	page, err := validatePage(args.Page, perPage, totalCount)
	if err != nil {
		return err
	}

	skipCount := validateSkipCount(page, perPage)
	pageSize := tmmath.MinInt(perPage, totalCount-skipCount)

	apiResults := make([]*ctypes.ResultBlock, 0, pageSize)
	for i := skipCount; i < skipCount+pageSize; i++ {
		block := s.vm.blockStore.LoadBlock(results[i])
		if block != nil {
			blockMeta := s.vm.blockStore.LoadBlockMeta(block.Height)
			if blockMeta != nil {
				apiResults = append(apiResults, &ctypes.ResultBlock{
					Block:   block,
					BlockID: blockMeta.BlockID,
				})
			}
		}
	}

	reply.Blocks = apiResults
	reply.TotalCount = totalCount

func (s *LocalService) BlockchainInfo(
	_ *http.Request,
	args *BlockchainInfoArgs,
	reply *ctypes.ResultBlockchainInfo,
) error {
	// maximum 20 block metas
	const limit int64 = 20
	var err error
	args.MinHeight, args.MaxHeight, err = core.FilterMinMax(
		s.vm.blockStore.Base(),
		s.vm.blockStore.Height(),
		args.MinHeight,
		args.MaxHeight,
		limit)
	if err != nil {
		return err
	}
	s.vm.tmLogger.Debug("BlockchainInfoHandler", "maxHeight", args.MaxHeight, "minHeight", args.MinHeight)

	var blockMetas []*types.BlockMeta
	for height := args.MaxHeight; height >= args.MinHeight; height-- {
		blockMeta := s.vm.blockStore.LoadBlockMeta(height)
		blockMetas = append(blockMetas, blockMeta)
	}

	reply.LastHeight = s.vm.blockStore.Height()
	reply.BlockMetas = blockMetas

	return nil
}

func (s *LocalService) Genesis(_ *http.Request, _ *struct{}, reply *ctypes.ResultGenesis) error {
	if len(s.vm.genChunks) > 1 {
		return errors.New("genesis response is large, please use the genesis_chunked API instead")
	}

	reply.Genesis = s.vm.genesis

	return nil
}

func (s *LocalService) GenesisChunked(_ *http.Request, args *GenesisChunkedArgs, reply *ctypes.ResultGenesisChunk) error {
	if s.vm.genChunks == nil {
		return fmt.Errorf("service configuration error, genesis chunks are not initialized")
	}

	if len(s.vm.genChunks) == 0 {
		return fmt.Errorf("service configuration error, there are no chunks")
	}

	id := int(args.Chunk)

	if id > len(s.vm.genChunks)-1 {
		return fmt.Errorf("there are %d chunks, %d is invalid", len(s.vm.genChunks)-1, id)
	}

	reply.TotalChunks = len(s.vm.genChunks)
	reply.ChunkNumber = id
	reply.Data = s.vm.genChunks[id]
	return nil
}

func (s *LocalService) Status(_ *http.Request, args *StatusArgs, reply *ctypes.ResultStatus) error {
	var (
		earliestBlockHeight   int64
		earliestBlockHash     tmbytes.HexBytes
		earliestAppHash       tmbytes.HexBytes
		earliestBlockTimeNano int64
	)

	if earliestBlockMeta := s.vm.blockStore.LoadBaseMeta(); earliestBlockMeta != nil {
		earliestBlockHeight = earliestBlockMeta.Header.Height
		earliestAppHash = earliestBlockMeta.Header.AppHash
		earliestBlockHash = earliestBlockMeta.BlockID.Hash
		earliestBlockTimeNano = earliestBlockMeta.Header.Time.UnixNano()
	}

	var (
		latestBlockHash     tmbytes.HexBytes
		latestAppHash       tmbytes.HexBytes
		latestBlockTimeNano int64

		latestHeight = s.vm.blockStore.Height()
	)

	if latestHeight != 0 {
		if latestBlockMeta := s.vm.blockStore.LoadBlockMeta(latestHeight); latestBlockMeta != nil {
			latestBlockHash = latestBlockMeta.BlockID.Hash
			latestAppHash = latestBlockMeta.Header.AppHash
			latestBlockTimeNano = latestBlockMeta.Header.Time.UnixNano()
		}
	}

	reply.NodeInfo = p2p.DefaultNodeInfo{
		DefaultNodeID: p2p.ID(s.vm.ctx.NodeID.String()),
		Network:       fmt.Sprintf("%d", s.vm.ctx.NetworkID),
	}
	reply.SyncInfo = ctypes.SyncInfo{
		LatestBlockHash:     latestBlockHash,
		LatestAppHash:       latestAppHash,
		LatestBlockHeight:   latestHeight,
		LatestBlockTime:     time.Unix(0, latestBlockTimeNano),
		EarliestBlockHash:   earliestBlockHash,
		EarliestAppHash:     earliestAppHash,
		EarliestBlockHeight: earliestBlockHeight,
		EarliestBlockTime:   time.Unix(0, earliestBlockTimeNano),
	}
	return nil
}

// ToDo: do we need to implement this method?
// ToDo: we don't have access to p2p info
func (s *LocalService) NetInfo(_ *http.Request, args *NetInfoArgs, reply *ctypes.ResultNetInfo) error {
	return nil
}

// ToDo: do we need to implement this method?
// ToDo: we don't have access to p2p info
func (s *LocalService) DumpConsensusState(_ *http.Request, args *DumpConsensusStateArgs, reply *ctypes.ResultDumpConsensusState) error {
	return nil
}

// ToDo: do we need to implement this method?
// ToDo: we don't have consensus
func (s *LocalService) ConsensusState(_ *http.Request, args *ConsensusStateArgs, reply *ctypes.ResultConsensusState) error {
	return nil
}

// ToDo: do we need to implement this method?
// ToDo: we don't have consensus
func (s *LocalService) ConsensusParams(_ *http.Request, args *ConsensusParamsArgs, reply *ctypes.ResultConsensusParams) error {
	return nil
}

func (s *LocalService) Health(_ *http.Request, args *HealthArgs, reply *ctypes.ResultHealth) error {
	*reply = ctypes.ResultHealth{}
	return nil
}

func (s *LocalService) UnconfirmedTxs(_ *http.Request, args *UnconfirmedTxsArgs, reply *ctypes.ResultUnconfirmedTxs) error {
	limit := validatePerPage(args.Limit)
	txs := s.vm.mempool.ReapMaxTxs(limit)
	reply.Count = len(txs)
	reply.Total = s.vm.mempool.Size()
	reply.Txs = txs
	return nil
}

func (s *LocalService) NumUnconfirmedTxs(_ *http.Request, args *NumUnconfirmedTxsArgs, reply *ctypes.ResultUnconfirmedTxs) error {
	reply.Count = s.vm.mempool.Size()
	reply.Total = s.vm.mempool.Size()
	reply.TotalBytes = s.vm.mempool.TxsBytes()
	return nil
}

func (s *LocalService) CheckTx(_ *http.Request, args *CheckTxArgs, reply *ctypes.ResultCheckTx) error {
	res, err := s.vm.proxyApp.Mempool().CheckTxSync(abci.RequestCheckTx{Tx: args.Tx})
	if err != nil {
		return err
	}
	reply.ResponseCheckTx = *res
	return nil
}
