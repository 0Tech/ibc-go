package types_test

import (
	"fmt"
	"time"

	tmtypes "github.com/Finschia/ostracon/types"
	"github.com/tendermint/tendermint/crypto/tmhash"

	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v4/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/cosmos/ibc-go/v4/modules/light-clients/07-tendermint/types"
	ibctestingmock "github.com/cosmos/ibc-go/v4/testing/mock"
)

func (suite *TendermintTestSuite) TestCheckMisbehaviourAndUpdateState() {
	altPrivVal := ibctestingmock.NewPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)

	altVal := tmtypes.NewValidator(altPubKey, 4)

	// Create alternative validator set with only altVal
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	// Create bothValSet with both suite validator and altVal
	bothValSet, bothSigners := getBothSigners(suite, altVal, altPrivVal)
	bothValsHash := bothValSet.Hash()

	altSigners := getAltSigners(altVal, altPrivVal)

	heightMinus1 := clienttypes.NewHeight(height.RevisionNumber, height.RevisionHeight-1)
	heightMinus3 := clienttypes.NewHeight(height.RevisionNumber, height.RevisionHeight-3)

	testCases := []struct {
		name            string
		clientState     exported.ClientState
		consensusState1 exported.ConsensusState
		height1         clienttypes.Height
		consensusState2 exported.ConsensusState
		height2         clienttypes.Height
		misbehaviour    exported.Misbehaviour
		timestamp       time.Time
		expPass         bool
	}{
		{
			"valid fork misbehaviour",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"valid time misbehaviour",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+3), height, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"valid time misbehaviour header 1 stricly less than header 2",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+3), height, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now.Add(time.Hour), bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"valid misbehavior at height greater than last consensusState",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), heightMinus1, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), heightMinus1, suite.now.Add(time.Minute), bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"valid misbehaviour with different trusted heights",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			heightMinus3,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), heightMinus1, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), heightMinus3, suite.now.Add(time.Minute), bothValSet, bothValSet, suite.valSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"valid misbehaviour at a previous revision",
			types.NewClientState(chainIDRevision1, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, clienttypes.NewHeight(1, 1), commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			heightMinus3,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainIDRevision0, int64(height.RevisionHeight+1), heightMinus1, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainIDRevision0, int64(height.RevisionHeight+1), heightMinus3, suite.now.Add(time.Minute), bothValSet, bothValSet, suite.valSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"valid misbehaviour at a future revision",
			types.NewClientState(chainIDRevision0, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			heightMinus3,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainIDRevision0, 3, heightMinus1, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainIDRevision0, 3, heightMinus3, suite.now.Add(time.Minute), bothValSet, bothValSet, suite.valSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"valid misbehaviour with trusted heights at a previous revision",
			types.NewClientState(chainIDRevision1, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, clienttypes.NewHeight(1, 1), commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			heightMinus3,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainIDRevision1, 1, heightMinus1, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainIDRevision1, 1, heightMinus3, suite.now.Add(time.Minute), bothValSet, bothValSet, suite.valSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"consensus state's valset hash different from misbehaviour should still pass",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now, bothValSet, bothValSet, suite.valSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now.Add(time.Minute), bothValSet, bothValSet, suite.valSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"invalid fork misbehaviour: identical headers",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"invalid time misbehaviour: monotonically increasing time",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+3), height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"invalid misbehavior misbehaviour from different chain",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader("ethermint", int64(height.RevisionHeight+1), height, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader("ethermint", int64(height.RevisionHeight+1), height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"invalid misbehavior misbehaviour with trusted height different from trusted consensus state",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			heightMinus3,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), heightMinus1, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now.Add(time.Minute), bothValSet, bothValSet, suite.valSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"invalid misbehavior misbehaviour with trusted validators different from trusted consensus state",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			heightMinus3,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), heightMinus1, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), heightMinus3, suite.now.Add(time.Minute), bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"already frozen client state",
			&types.ClientState{FrozenHeight: clienttypes.NewHeight(0, 1)},
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"trusted consensus state does not exist",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			nil, // consensus state for trusted height - 1 does not exist in store
			clienttypes.Height{},
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), heightMinus1, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"invalid tendermint misbehaviour",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			nil,
			suite.now,
			false,
		},
		{
			"provided height > header height",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), heightMinus1, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), heightMinus1, suite.now.Add(time.Minute), bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"trusting period expired",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(time.Time{}, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), heightMinus1, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now.Add(trustingPeriod),
			false,
		},
		{
			"trusted validators is incorrect for given consensus state",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now, bothValSet, bothValSet, suite.valSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now.Add(time.Minute), bothValSet, bothValSet, suite.valSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"first valset has too much change",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now, altValSet, altValSet, bothValSet, altSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothValSet, bothSigners),
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"second valset has too much change",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now, bothValSet, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now.Add(time.Minute), altValSet, altValSet, bothValSet, altSigners),
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"both valsets have too much change",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now, altValSet, altValSet, bothValSet, altSigners),
				Header2:  suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight+1), height, suite.now.Add(time.Minute), altValSet, altValSet, bothValSet, altSigners),
				ClientId: chainID,
			},
			suite.now,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case: %s", tc.name), func() {
			// reset suite to create fresh application state
			suite.SetupTest()

			// Set current timestamp in context
			ctx := suite.chainA.GetContext().WithBlockTime(tc.timestamp)

			// Set trusted consensus states in client store

			if tc.consensusState1 != nil {
				suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(ctx, clientID, tc.height1, tc.consensusState1)
			}
			if tc.consensusState2 != nil {
				suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(ctx, clientID, tc.height2, tc.consensusState2)
			}

			clientState, err := tc.clientState.CheckMisbehaviourAndUpdateState(
				ctx,
				suite.chainA.App.AppCodec(),
				suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(ctx, clientID), // pass in clientID prefixed clientStore
				tc.misbehaviour,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
				suite.Require().NotNil(clientState, "valid test case %d failed: %s", i, tc.name)
				suite.Require().True(!clientState.(*types.ClientState).FrozenHeight.IsZero(), "valid test case %d failed: %s", i, tc.name)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
				suite.Require().Nil(clientState, "invalid test case %d passed: %s", i, tc.name)
			}
		})
	}
}
