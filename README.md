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


