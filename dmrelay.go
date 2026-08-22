// mautrix-discord - A Matrix-Discord puppeting bridge.
//
// DM relay state, stored in a table this fork owns.
//
// DELIBERATELY NOT A SCHEMA MIGRATION. Registering this as an upgrade would bump
// the bridge's schema version past what upstream mautrix-discord knows (24), and
// whether upstream then refuses to start is a question about dbutil's version
// handling that is not worth betting a downgrade path on. Creating the table
// outside the migration system leaves `version` untouched at 24, so upstream sees
// a database it fully recognises and simply never queries this table. Going back
// to the unforked image is then a pure image swap.
//
// For the same reason there is no foreign key to `portal`: upstream must never be
// able to trip over this table, including when it deletes portals.

package main

import (
	"database/sql"

	"maunium.net/go/mautrix/id"

	"go.mau.fi/mautrix-discord/database"
)

// Table name is prefixed so it cannot collide with a table upstream might add later.
const createDMRelayTable = `
	CREATE TABLE IF NOT EXISTS lhns_dm_relay (
		dcid       TEXT NOT NULL,
		receiver   TEXT NOT NULL,
		relay_user TEXT NOT NULL,
		PRIMARY KEY (dcid, receiver)
	)
`

// initDMRelay creates the table if needed and loads it into memory. Called once
// at startup: the table is tiny (one row per relaying DM) and reading it per
// message would put a query in the hot path for no reason.
func (br *DiscordBridge) initDMRelay() {
	if _, err := br.DB.Exec(createDMRelayTable); err != nil {
		br.ZLog.Err(err).Msg("Failed to create DM relay table")
		return
	}
	rows, err := br.DB.Query(`SELECT dcid, receiver, relay_user FROM lhns_dm_relay`)
	if err != nil {
		br.ZLog.Err(err).Msg("Failed to load DM relay table")
		return
	}
	defer rows.Close()

	loaded := make(map[database.PortalKey]id.UserID)
	for rows.Next() {
		var key database.PortalKey
		var relayUser string
		if err = rows.Scan(&key.ChannelID, &key.Receiver, &relayUser); err != nil {
			br.ZLog.Err(err).Msg("Failed to scan DM relay row")
			return
		}
		loaded[key] = id.UserID(relayUser)
	}
	br.dmRelayLock.Lock()
	br.dmRelay = loaded
	br.dmRelayLock.Unlock()
	br.ZLog.Debug().Int("count", len(loaded)).Msg("Loaded DM relay settings")
}

// GetDMRelayUser returns the Matrix user relaying for a DM portal, or "" if none.
func (br *DiscordBridge) GetDMRelayUser(key database.PortalKey) id.UserID {
	br.dmRelayLock.RLock()
	defer br.dmRelayLock.RUnlock()
	return br.dmRelay[key]
}

// SetDMRelayUser marks a DM portal as relaying through the given Matrix user.
func (br *DiscordBridge) SetDMRelayUser(key database.PortalKey, user id.UserID) error {
	_, err := br.DB.Exec(`
		INSERT INTO lhns_dm_relay (dcid, receiver, relay_user) VALUES ($1, $2, $3)
		ON CONFLICT (dcid, receiver) DO UPDATE SET relay_user=excluded.relay_user
	`, key.ChannelID, key.Receiver, user.String())
	if err != nil {
		return err
	}
	br.dmRelayLock.Lock()
	if br.dmRelay == nil {
		br.dmRelay = make(map[database.PortalKey]id.UserID)
	}
	br.dmRelay[key] = user
	br.dmRelayLock.Unlock()
	return nil
}

// ClearDMRelayUser turns relaying off for a DM portal. Also called when a portal
// is deleted, so a rebuilt portal does not silently inherit the old setting.
func (br *DiscordBridge) ClearDMRelayUser(key database.PortalKey) error {
	_, err := br.DB.Exec(`DELETE FROM lhns_dm_relay WHERE dcid=$1 AND receiver=$2`,
		key.ChannelID, key.Receiver)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	br.dmRelayLock.Lock()
	delete(br.dmRelay, key)
	br.dmRelayLock.Unlock()
	return nil
}
