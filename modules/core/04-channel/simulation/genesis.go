package simulation

import (
	"math/rand"

	simtypes "github.com/Finschia/finschia-sdk/types/simulation"

	"github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
)

// GenChannelGenesis returns the default channel genesis state.
func GenChannelGenesis(_ *rand.Rand, _ []simtypes.Account) types.GenesisState {
	return types.DefaultGenesisState()
}
