// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package usercontentupdatelambda

import (
	"sync"

	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/ddb"
	dbeventlambda "github.com/palchukovsky/ss/lambda/dbevent"
)

type Updater interface{ SpawnUpdate() }

type UpdaterFactory = func(
	user ss.UserID,
	request dbeventlambda.Request,
	db ddb.Client,
	finishBarrier *sync.WaitGroup,
) Updater
