package main

import (
	"fmt"
	"time"
)

// la fonction Max retourne la valeur la plus grande entre les deux paramètres fournis
func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// définition d'une structure de Message
type Message struct {
	id       int  // l'id d'un noeud
	isLeader bool // si l'id envoyé est considéré comme un leader
}

// définition d'une structure de Noeud
type RingNode struct {
	inChannel  <-chan Message // la connexion entrante
	outChannel chan<- Message // la connexion sortante
	localId    int            // l'id du noeud
	leaderId   int            // l'id du leader
}

func NewRingNode(id int, inChannel <-chan Message, outChannel chan<- Message) *RingNode {
	return &RingNode{inChannel, outChannel, id, 0}
}

// code principal des noeuds
func (r *RingNode) Run(sync chan int) {
	// Initialisation: chaque Noeud envoie son id au suivant
	fmt.Println("From", r.localId, ": Sending ", r.localId)
	r.outChannel <- Message{r.localId, false}

	for {

		// on attend la synchronisation du main
		<-sync

		receivedMessage := <-r.inChannel // on réception le message
		fmt.Println("From", r.localId, ": received", receivedMessage.id, receivedMessage.isLeader)

		if r.leaderId != 0 {
			// si le leaderId à été modifié alors le leader à été identifié, on ne fait rien
			continue
		} else if receivedMessage.isLeader == false { // si le leader n'as pas encore été idenfié

			// si l'id recu est l'id du noeud courant
			if receivedMessage.id == r.localId {
				done++ // utilisé pour la synchronisation
				// si l'id recu est égal à celui du noeud courant alors on viens d'identifier le leader
				r.leaderId = receivedMessage.id // on met à jour la valeur de l'id du leader
				fmt.Println("From", r.localId, ": Leader found at", r.leaderId)
				r.outChannel <- Message{r.leaderId, true} // on envoie un message contenant l'id du leader et on met à jour la valeur isLeader à true
				close(r.outChannel)                       // on ferme le channel
			} else { // si l'id recu n'est pas l'id courant
				max := Max(r.localId, receivedMessage.id)                                                               // on compare les deux id et on renvoie le plus grand des deux
				fmt.Println("From", r.localId, ": compared (", r.localId, "-", receivedMessage.id, ") | sending:", max) // on affiche un message informatif
				r.outChannel <- Message{max, false}                                                                     // on renvoie le plus grand id des deux
			}

		} else if receivedMessage.isLeader == true { // si le leader est identifié (peut etre remplacé par un else standard)
			done++                          // utilisé pour la synchronisation
			r.leaderId = receivedMessage.id // on met à jour la valeur de leaderId avec l'id recu
			fmt.Println("From", r.localId, ": Leader found at", r.leaderId, "| spreading message and closing channel")
			r.outChannel <- receivedMessage // on envoie le message recu au noeud suivant (pas besoin de le modifier)
			close(r.outChannel)             // on ferme la connexion sortante du noeud courant

		}
	}
}

const N = 5

var done = 0

func main() {

	// on créé les channels utilisé pour synchroniser les noeuds
	syncs := make([]chan int, N)

	for i := 0; i < N; i++ {
		syncs[i] = make(chan int, 1)
	}

	// on créé les channels de connexion entre les noeud
	outChannels := make([]chan Message, N)

	for i := 0; i < N; i++ {
		outChannels[i] = make(chan Message, 1)
	}

	// on créé les noeuds de l'anneau
	nodes := make([]*RingNode, N)
	for i := 0; i < N; i++ {
		next := (i + 1) % N
		nodes[i] = NewRingNode(i, outChannels[i], outChannels[next])
		go nodes[i].Run(syncs[i]) // on lance les noeuds avec un channel de synchronisation
	}

	for {
		// This is not so nice
		if done == N {
			return
		}
		time.Sleep(time.Second)
		fmt.Println()
		for i := 0; i < N; i++ {
			syncs[i] <- 0
		}
	}
}
