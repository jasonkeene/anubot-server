// Code generated by "stringer -type TwitchUser"; DO NOT EDIT

package store

import "fmt"

const _TwitchUser_name = "StreamerBot"

var _TwitchUser_index = [...]uint8{0, 8, 11}

func (i TwitchUser) String() string {
	if i < 0 || i >= TwitchUser(len(_TwitchUser_index)-1) {
		return fmt.Sprintf("TwitchUser(%d)", i)
	}
	return _TwitchUser_name[_TwitchUser_index[i]:_TwitchUser_index[i+1]]
}
