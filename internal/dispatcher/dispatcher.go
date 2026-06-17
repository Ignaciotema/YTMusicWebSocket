package dispatcher

type Player interface {
	PlayPause()
	Previous()
	Next()
}

type Dispatcher struct {
	player Player
}

func New(player Player) *Dispatcher {
	return &Dispatcher{
		player: player,
	}
}

func (d *Dispatcher) DispatchCommand(command string) {

	//chequea si tiene que mandar

	switch command {
	case "playPause":
		d.player.PlayPause()
	case "previous":
		d.player.Previous()
	case "next":
		d.player.Next()
	}
}
