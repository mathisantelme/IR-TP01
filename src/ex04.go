package main

import (
	"fmt"
	"sync"
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
	maxId      int            // l'id maximal enregistré par le noeud
	leaderId   int            // l'id du leader
}

func NewRingNode(id int, inChannel <-chan Message, outChannel chan<- Message) *RingNode {
	return &RingNode{inChannel, outChannel, id, id, 0}
}

// code principal des noeuds
func (r *RingNode) Run(wg *sync.WaitGroup) {
	// Done va etre executée un fois la boucle terminée (le inChannel est fermée)
	defer wg.Done()

	// Initialisation: chaque Noeud envoie son id au suivant
	fmt.Println("From", r.localId, ": Sending ", r.localId)
	r.outChannel <- Message{r.localId, false}

	for receivedMessage := range r.inChannel {
		fmt.Println("From", r.localId, ": received", receivedMessage.id, receivedMessage.isLeader)

		if r.leaderId != 0 {
			// si le leaderId à été modifié alors le leader à été identifié, on ne fait rien
			continue
		} else if receivedMessage.isLeader == false { // si le leader n'as pas encore été idenfié

			// si l'id recu est l'id du noeud courant
			if receivedMessage.id == r.localId {
				// si l'id recu est égal à celui du noeud courant alors on viens d'identifier le leader
				r.leaderId = receivedMessage.id // on met à jour la valeur de l'id du leader
				fmt.Println("From", r.localId, ": Leader found at", r.leaderId)
				r.outChannel <- Message{r.leaderId, true} // on envoie un message contenant l'id du leader et on met à jour la valeur isLeader à true
				close(r.outChannel)                       // on ferme le channel
			} else { // si l'id recu n'est pas l'id courant
				// si l'id recu est plus grand que le plus grand Id envoyé
				if receivedMessage.id > r.maxId {
					max := Max(r.localId, receivedMessage.id)                                                               // on compare les deux id et on renvoie le plus grand des deux
					fmt.Println("From", r.localId, ": compared (", r.localId, "-", receivedMessage.id, ") | sending:", max) // on affiche un message informatif
					r.outChannel <- Message{max, false}
				} // on renvoie le plus grand id des deux
			}

		} else if receivedMessage.isLeader == true { // si le leader est identifié (peut etre remplacé par un else standard)

			r.leaderId = receivedMessage.id // on met à jour la valeur de leaderId avec l'id recu
			fmt.Println("From", r.localId, ": Leader found at", r.leaderId, "| spreading message and closing channel")
			r.outChannel <- receivedMessage // on envoie le message recu au noeud suivant (pas besoin de le modifier)
			close(r.outChannel)             // on ferme la connexion sortante du noeud courant

		}
	}
}

func main() {
	out1 := make(chan Message, 1)
	out2 := make(chan Message, 1)
	out3 := make(chan Message, 1)

	node1 := NewRingNode(1, out3, out1)
	node2 := NewRingNode(2, out1, out2)
	node3 := NewRingNode(3, out2, out3)

	var wg sync.WaitGroup
	wg.Add(3)

	go node1.Run(&wg)
	go node2.Run(&wg)
	go node3.Run(&wg)

	wg.Wait()
}
