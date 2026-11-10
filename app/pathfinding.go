package main

import (
	"cmp"
	"math"
	"slices"
)

// Point repräsentiert eine Koordinate im 2D-Gitter
type Point struct {
	X, Y int
}

// Node bündelt alle Suchdaten eines Punktes
type Node struct {
	G, F      int
	Parent    Point
	HasParent bool // false beim Startknoten
}

// Solver hält den Zustand der A*-Suche. Die Felder OpenSet, ClosedSet,
// Current, Path und Done sind öffentlich, damit der Renderer direkt darauf
// zugreifen kann – es gibt keinen Step-Zwischenschritt mehr.
type Solver struct {
	OpenSet   []Point        // noch offene Punkte
	ClosedSet map[Point]bool // bereits verarbeitete Punkte
	Current   Point          // gerade verarbeiteter Punkt
	Path      []Point        // kompletter Endpfad; nil bis das Ziel erreicht ist
	Done      bool           // Suche abgeschlossen

	walls  map[Point]bool // besetzte Felder (true = Wand)
	width  int            // Gitterbreite
	height int            // Gitterhoehe
	goal   Point
	nodes  map[Point]*Node // Suchdaten je Punkt (intern)
}

// Die 4 orthogonalen Bewegungsrichtungen
var directions = []Point{
	{X: 0, Y: -1}, // Hoch
	{X: 0, Y: 1},  // Runter
	{X: -1, Y: 0}, // Links
	{X: 1, Y: 0},  // Rechts
}

// Manhattan-Distanz als Heuristik (4 Richtungen)
func heuristic(a, b Point) int {
	absX := a.X - b.X
	if absX < 0 {
		absX = -absX
	}
	absY := a.Y - b.Y
	if absY < 0 {
		absY = -absY
	}
	return absX + absY
}

// Prüft, ob ein Punkt im Gitter liegt und nicht blockiert ist
func isValid(p Point, width, height int, walls map[Point]bool) bool {
	return p.X >= 0 && p.X < width && p.Y >= 0 && p.Y < height && !walls[p]
}

// NewSolver legt den Suchzustand an.
// Alle begehbaren Punkte werden vorab mit "unendlich" (math.MaxInt) angelegt,
// damit in der Schleife kein Sonderfall für neue Knoten nötig ist.
func NewSolver(walls map[Point]bool, width, height int, start, goal Point) *Solver {
	nodes := map[Point]*Node{}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := Point{X: x, Y: y}
			if !walls[p] { // begehbar
				nodes[p] = &Node{G: math.MaxInt, F: math.MaxInt}
			}
		}
	}
	// Der Startknoten bekommt direkt seine echten Werte
	nodes[start] = &Node{G: 0, F: heuristic(start, goal)}

	return &Solver{
		walls:     walls,
		width:     width,
		height:    height,
		goal:      goal,
		OpenSet:   []Point{start},
		ClosedSet: map[Point]bool{},
		nodes:     nodes,
	}
}

// Next führt genau einen Expansionsschritt der Suche aus. Der Renderer
// liest die öffentlichen Felder danach direkt aus.
func (s *Solver) Next() {
	if s.Done {
		return
	}

	// Aufsteigend nach F sortieren: kleinstes F an Index 0.
	// Jeder Punkt im OpenSet hat garantiert einen Eintrag in nodes.
	slices.SortFunc(s.OpenSet, func(a, b Point) int {
		return cmp.Compare(s.nodes[a].F, s.nodes[b].F)
	})

	s.Current = s.OpenSet[0]

	// Ziel erreicht?
	if s.Current == s.goal {
		s.Path = reconstructPath(s.nodes, s.Current)
		s.Done = true
		return
	}

	s.OpenSet = s.OpenSet[1:]
	s.ClosedSet[s.Current] = true

	currentG := s.nodes[s.Current].G

	for _, dir := range directions {
		neighbor := Point{X: s.Current.X + dir.X, Y: s.Current.Y + dir.Y}
		if !isValid(neighbor, s.width, s.height, s.walls) {
			continue
		}
		// Bereits endgültig verarbeitet -> überspringen
		if s.ClosedSet[neighbor] {
			continue
		}

		// Ein Schritt kostet immer 1
		tentativeG := currentG + 1
		// Existiert garantiert: alle begehbaren Punkte wurden vorab angelegt
		n := s.nodes[neighbor]

		// Kürzerer Weg gefunden als bisher bekannt? ("unendlich" = erster Fund)
		if tentativeG < n.G {
			n.G = tentativeG
			n.F = tentativeG + heuristic(neighbor, s.goal)
			n.Parent = s.Current
			n.HasParent = true

			if !slices.Contains(s.OpenSet, neighbor) {
				s.OpenSet = append(s.OpenSet, neighbor)
			}
		}
	}

	if len(s.OpenSet) == 0 {
		s.Done = true // Kein Pfad gefunden
	}
}

// CurrentPath liefert den anzuzeigenden Pfad: während der Suche den
// bisherigen Teilpfad Start -> Current, nach Abschluss den kompletten
// Endpfad (nil, wenn keine Lösung existiert).
func (s *Solver) CurrentPath() []Point {
	if s.Done {
		return s.Path
	}
	return reconstructPath(s.nodes, s.Current)
}

// Läuft über die Parent-Verweise vom Ziel zurück zum Start
func reconstructPath(nodes map[Point]*Node, current Point) []Point {
	path := []Point{current}
	for {
		n := nodes[current]
		if !n.HasParent {
			break
		}
		current = n.Parent
		path = append(path, current)
	}
	// Pfad wurde rückwärts gebaut, also umdrehen
	slices.Reverse(path)
	return path
}

// FindPath sucht per A* den kürzesten Weg von start nach goal im Grid.
// 0 = begehbar, 1 = Wand. Liefert nil, wenn kein Pfad existiert.
// Bequemlichkeits-Wrapper um den Solver: führt die Suche vollständig aus.
func FindPath(walls map[Point]bool, width, height int, start, goal Point) []Point {
	s := NewSolver(walls, width, height, start, goal)
	for !s.Done {
		s.Next()
	}
	return s.Path
}
