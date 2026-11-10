package main

import (
	"fmt"
	"math/rand/v2"
	"os"

	"path3d/lib3d"
)

// ====================================================================
// path3d - Grundgeruest fuer spaetere Pfadsuche.
//
// Ein Bodengitter aus 30x30 Feldern. Auf den ersten beiden Zeilen steht je
// ein Wuerfel, der genau ein Feld bedeckt. Die Wuerfel werden direkt per
// (row, col)-Schleife platziert - es gibt keine Grid-Datenstruktur.
//
// Konvention: col zeigt in X-Richtung, row in Z-Richtung. Das Gitter liegt
// zentriert um den Ursprung in der XZ-Ebene (y=0).
// ====================================================================

const (
	edgeCells  = 40                            // Anker: Felder pro Achse
	cellSize   = 1.0                           // Kantenlaenge eines Feldes
	edgeLength = float64(edgeCells) * cellSize // Kantenlaenge des Gitters (pro Achse)
	cubeSize   = 0.9                           // Wuerfelkante (etwas kleiner als das Feld, damit klar bleibt, dass er EIN Feld bedeckt)
	filledRows = 2                             // Anzahl der ersten Zeilen, die mit Wuerfeln belegt werden
)

// cellCenter rechnet Weltkoordinaten der Feldmitte fuer ein Feld (row, col).
// col -> X-Achse, row -> Z-Achse, Feldmitte bekommt die +0.5-Verschiebung.
func cellCenter(row, col int) (x float64, z float64) {
	half := edgeLength / 2
	x = -half + (float64(col)+0.5)*cellSize
	z = -half + (float64(row)+0.5)*cellSize
	return x, z
}

// randomGrid erzeugt eine Map mit besetzten Feldern (true = Wand) fuer ein
// width x height-Grid mit zufällig verteilten Wänden.
// math/rand ist ab Go 1.20 automatisch pro Programmstart anders seedet –
// jede Ausführung ergibt also ein anderes Gitter.
func randomGrid(width, height int, wallChance float64) map[Point]bool {
	walls := make(map[Point]bool)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if rand.Float64() < wallChance {
				walls[Point{X: x, Y: y}] = true
			}
		}
	}
	return walls
}

func drawCubes(instances []lib3d.Instance, solver *Solver) []lib3d.Instance {

	// Besetzte Grid-Felder (Waende) in rot
	cubeCol := lib3d.ParseHex("#ff0000")
	for p := range solver.walls {
		x, z := cellCenter(p.X, p.Y)
		instances = append(instances, lib3d.Instance{
			Model: lib3d.TranslateMatrix(x, cubeSize/2, z),
			Color: cubeCol,
		})
	}

	// Bereits verarbeitete Punkte (orange)
	cubeCol = lib3d.ParseHex("#ffcc33")
	for p := range solver.ClosedSet {
		x, z := cellCenter(p.X, p.Y)
		instances = append(instances, lib3d.Instance{
			Model: lib3d.TranslateMatrix(x, cubeSize/2, z),
			Color: cubeCol,
		})
	}

	// Noch offene Punkte (hellblau)
	cubeCol = lib3d.ParseHex("#87CEEB")
	for _, p := range solver.OpenSet {
		x, z := cellCenter(p.X, p.Y)
		instances = append(instances, lib3d.Instance{
			Model: lib3d.TranslateMatrix(x, cubeSize/2, z),
			Color: cubeCol,
		})
	}

	// Bisher gefundener Pfad (grün)
	cubeCol = lib3d.ParseHex("#87eba2")
	for _, p := range solver.CurrentPath() {
		x, z := cellCenter(p.X, p.Y)
		instances = append(instances, lib3d.Instance{
			Model: lib3d.TranslateMatrix(x, cubeSize/2, z),
			Color: cubeCol,
		})
	}

	return instances
}

func main() {
	walls := randomGrid(int(edgeCells), edgeCells, 0.3)
	start := Point{X: 0, Y: 0}
	goal := Point{X: edgeCells - 1, Y: edgeCells - 1}
	delete(walls, start) // Start darf keine Wand sein
	delete(walls, goal)  // Ziel darf keine Wand sein

	solver := NewSolver(walls, int(edgeCells), edgeCells, start, goal)

	scene := lib3d.NewScene()
	scene.Cam.Pos = lib3d.Vec3{X: 17, Y: 24, Z: 32}
	scene.Cam.Target = lib3d.Vec3{X: 0, Y: 0, Z: 0}
	scene.Cam.Up = lib3d.Vec3{X: 0, Y: 1, Z: 0}
	scene.Cam.Far = 500

	// Meshes EINMAL erzeugen und wiederverwenden.
	cubeMesh := lib3d.MakeCubeMesh(cubeSize, cubeSize, cubeSize)
	gridMesh := lib3d.MakeGridMesh(edgeLength, edgeCells)

	gridStyle := lib3d.ShapeStyle{
		Stroke:    lib3d.Hex("#777774"),
		LineWidth: 1,
	}
	cubeStyle := lib3d.ShapeStyle{
		Stroke:    lib3d.Hex("#6b5200"),
		Fill:      lib3d.Hex("#ffffff"), // wird pro Instanz ueberschrieben
		LineWidth: 1,
		Mode:      lib3d.StyleBoth,
	}

	instances := make([]lib3d.Instance, 0, edgeCells*edgeCells)

	/* Loop */
	draw := func() {
		if !solver.Done {
			solver.Next()
			if solver.Done {
				if path := solver.CurrentPath(); path != nil {
					fmt.Printf("Pfad gefunden (%d Schritte): ", len(path))
					for _, p := range path {
						fmt.Printf("[%d, %d] ", p.X, p.Y)
					}
					fmt.Println()
				} else {
					fmt.Println("Kein Pfad möglich.")
				}
			}
		}

		lib3d.BeginScene(scene)

		// Bodengitter (XZ-Ebene bei y=0) aus dem geteilten Mesh.
		lib3d.DrawMesh(gridMesh, gridStyle)

		// Alle Wuerfel in EINEM Instancing-Draw-Call.
		instances = instances[:0]
		instances = drawCubes(instances, solver)

		lib3d.DrawMeshInstanced(cubeMesh, cubeStyle, instances)
	}

	/* Start */
	if err := lib3d.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	lib3d.SetBatchSort(true)
	lib3d.StartAnimation(draw)
}
