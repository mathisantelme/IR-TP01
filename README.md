# Informatique Répartie - TP01

L’objectif de cette séance est de reprendre l’exercice sur l’anneau à jeton de la première semaine et
d’utiliser un principe similaire pour résoudre le problème de l’élection de leader sur un anneau.
<br></br>
Le principe est simple :

- On dispose d’un anneau de **N** sommets (**N** est inconnu de tous)
- Chaque sommet a un identifiant (entre 1 et **N**)
- Un sommet ne peut envoyer un message qu’au sommet suivant et recevoir un message que
du sommet précédent.

L’objectif est qu’à la fin de l’algorithme chaque sommet connaisse l’identifiant le plus élevé sur
l’anneau (n dans ce cas). Le sommet ayant cet identifiant est le leader.
<br></br>
L’algorithme est décomposé en deux phases : avant que le leader ne soit connu et après. Avant que
le leader ne soit connu :

- Initialement chaque sommet envoi son identifiant au sommet suivant dans l’anneau.
- A chaque réception de message un sommet va renvoyer le max entre son identifiant et
l’identifiant reçu.
- Si un sommet reçoit son propre identifiant c’est que son identifiant a fait tout le tour de l’anneau et était à chaque fois le max. Il est donc le leader.

Une fois le leader auto-identifié :

- Il doit envoyer un message "je suis le leader" au suivant sur l’anneau.
- Chaque sommet va propager le message du leader et l’algorithme se termine quand tous les
sommets savent qui est le leader.

Les deux différences avec l’exercice de la semaine 1 sont que le message envoyé contient une
information (c’est n’est pas juste un jeton) et surtout qu’il y a plusieurs messages en circulation au
même moment.

## Première Partie

1. A quel moment un sommet peut fermer sa connexion sortante ?
   
Un sommet peut fermer sa connexion sortante au moment ou il à envoyé le message indiquant l'identifiant du leader (si il ferme au moment de la réception de l'identifiant du leader, cela bloquerai les sommets suivants en attente du message).

2. Complétez le code fournit pour exécuter l'algorithme sur un anneau à trois sommets (seul le contenu de la boucle doit etre modifiée). Faire un affichage à chaque réception de message et afficher le nombre de messages envoyés.

Si on exécute le code fournit, on se rend compte que l'initialisation des Noeud s'effectue correctement (chaque noeud envoie son id à son voisin), et que les message sont correctement recus (les noeud recoivent les id des noeuds les précédant dans l'anneau). Seulement une fois ces deux étapes effectuées, les goroutines s'attendent mutuellement et créé un [deadlock](https://fr.wikipedia.org/wiki/Interblocage).
<br></br>
Afin d'éviter le deadlock, on doit permettre à chaque noeud de comparer l'identifiant recu et son propre id afin de renvoyer l'id le plus élevé. Pour cela on va compléter la boucle `for` présente dans la fonction `Run`.

```go
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
            close(r.outChannel) // on ferme le channel
        } else { // si l'id recu n'est pas l'id courant
            max := Max(r.localId, receivedMessage.id) // on compare les deux id et on renvoie le plus grand des deux
            fmt.Println("From", r.localId, ": compared (", r.localId, "-", receivedMessage.id, ") | sending:", max) // on affiche un message informatif
            r.outChannel <- Message{max, false} // on renvoie le plus grand id des deux
        }
    
    } else if receivedMessage.isLeader == true { // si le leader est identifié (peut etre remplacé par un else standard)
        
        // si l'id recu est égal à celui du noeud courant
        if receivedMessage.id != r.localId { // si l'id recu n'est pas l'id courant
            r.leaderId = receivedMessage.id // on met à jour la valeur de leaderId avec l'id recu
            fmt.Println("From", r.localId, ": Leader found at", r.leaderId, "| spreading message and closing channel")
            r.outChannel <- receivedMessage // on envoie le message recu au noeud suivant (pas besoin de le modifier)
            close(r.outChannel) // on ferme la connexion sortante du noeud courant
        }

    }
}
```

3. Généraliser tout le code pour créer un anneau avec **N** sommets (seul le main doit
être modifié). Vous pouvez tester des solutions dans lesquelles l’identifiant des sommets est
différent de leur position sur l’anneau.

Pour cela il suffit de créer des nouveaux noeuds dans le main et d'alterner leurs channel d'entrée et de sortie afin de changer leur position dans l'anneau.

```go
func main() {

    out1 := make(chan Message, 1)
    out2 := make(chan Message, 1)
    out3 := make(chan Message, 1)
    out4 := make(chan Message, 1)
    out5 := make(chan Message, 1)

	// On ajoute des nouveaux noeuds et on mélange leurs position sur l'anneau en modifiant les channels d'entrée et de sortie
    node1 := NewRingNode(1, out4, out1)
    node2 := NewRingNode(2, out5, out2)
    node3 := NewRingNode(3, out1, out3)
    node4 := NewRingNode(4, out2, out4)
    node5 := NewRingNode(5, out3, out5)

    var wg sync.WaitGroup
    wg.Add(3)

    go node1.Run(&wg)
    go node2.Run(&wg)
	go node3.Run(&wg)
	go node4.Run(&wg)
	go node5.Run(&wg)

    wg.Wait()
}
```