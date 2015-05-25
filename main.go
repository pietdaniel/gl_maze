package main

import (
	"container/list"
	//"fmt"
	"github.com/banthar/Go-SDL/sdl"
	"github.com/go-gl/gl/v2.1/gl"
	"math/rand"
	"time"
)

var view_rotx float64 = 0.0
var view_roty float64 = 90.0
var view_rotz float64 = 0.0
var view_z float64 = 0
var angle float64 = 0.0
var boxes []uint32 // for the drawing of walls FIXME should not be global imo
var width, height int32 = 20, 20

var DEBUG bool = false

//
// 3 x 3 cells
// 2 x 3 vert walls
// 2 x 3 horzontal walls
//       1,1     1,2
//  [1,1] | [1,2] | [1,3]
//   ---     ---     ---   1,1 1,2 1,3 H
//  [2,1] | [2,2] | [2,3]                { 2,1  2,2 V }
//   ---     ---     ---   2,1 2,2 2,3 H
//  [3,1] | [3,2] | [3,3]
//       3,1     3,2  V
//

// an array of cells where 1 is visited
var my_maze = make_2d_int(height, width)

// walls array where 1 is active 0 is torn down
var v_walls = make_2d_wall_s(height, width-1)
var h_walls = make_2d_wall_s(height-1, width)

func make_2d_int(h, w int32) [][]int32 {
	var param [][]int32 = make([][]int32, h)
	for i := range param {
		param[i] = make([]int32, w)
	}
	return param
}

func make_2d_wall_s(h, w int32) [][]*wall_s {
	var param [][]*wall_s = make([][]*wall_s, h)
	for i := range param {
		param[i] = make([]*wall_s, w)
	}
	return param
}

type pos struct {
	x, y int32
}

type wall_s struct {
	x, y, vertical, active int32
}

func new_wall_s(x, y, vertical, active int32) *wall_s {
	return &wall_s{x: x, y: y, vertical: vertical, active: active}
}

// allocates walls for the v/h_walls arrays
func fill_maze() {
	// all cells are not visited
	for _, x := range my_maze {
		for y := range x {
			x[y] = 0
		}
	}

	var x_ctr int32 = 0
	var y_ctr int32 = 0
	// all v_walls are up
	for _, x := range v_walls {
		for y := range x {
			x[y] = new_wall_s(x_ctr, y_ctr, 1, 1)
			y_ctr++
		}
		y_ctr = 0
		x_ctr++
	}
	x_ctr = 0
	y_ctr = 0
	// all h walls are up
	for _, x := range h_walls {
		for y := range x {
			x[y] = new_wall_s(x_ctr, y_ctr, 0, 1)
			y_ctr++
		}
		y_ctr = 0
		x_ctr++
	}
}

// retruns a rand maze similiar to prims
func rands() {
	rand.Seed(time.Now().UTC().UnixNano())
	for _, x := range v_walls {
		for y := range x {
			x[y].active = int32(rand.Intn(2))
		}
	}
	for _, x := range h_walls {
		for y := range x {
			x[y].active = int32(rand.Intn(2))
		}
	}
}

// container used in prims
var wall_list = list.New()

// if the wall is active, add to the list of walls
func push_wall(w *wall_s) {
	if w.active == 1 {
		wall_list.PushBack(w)
	}
}

/// adds the walls surrounding the give cell
func add_walls(cell pos) {
	var w, h int32
	w = cell.x
	h = cell.y

	var is_top = (cell.x == 0)
	var is_lft = (cell.y == 0)
	var is_bot = (cell.x == height-1)
	var is_rgt = (cell.y == width-1)

	if DEBUG {
		print("x ", cell.x, " y ", cell.y, "\n")
		print("t: ", is_top, " b: ", is_bot, " l: ", is_lft, " r: ", is_rgt, "\n")
		print(len(h_walls), " |[0] ", len(h_walls[0]), "\n")
		print(len(v_walls), " |[0] ", len(v_walls[0]), "\n")
	}

	if is_top {
		push_wall(h_walls[w][h])
	} else if is_bot {
		push_wall(h_walls[w-1][h])
	} else {
		push_wall(h_walls[w][h])
		push_wall(h_walls[w-1][h])
	}
	if is_lft {
		push_wall(v_walls[w][h])
	} else if is_rgt {
		push_wall(v_walls[w][h-1])
	} else {
		push_wall(v_walls[w][h])
		push_wall(v_walls[w][h-1])
	}
}

func prims() {
	//	Start with a grid full of walls.
	var cell pos
	var wall_pos pos
	var rand_wall int32
	var ctr int32 = 0

	//	Pick a cell, mark it as part of the maze. Add the walls of the cell to the wall list.
	rand.Seed(time.Now().UTC().UnixNano())
	cell = pos{int32(rand.Intn(int(height))), int32(rand.Intn(int(width)))}
	my_maze[cell.x][cell.y] = 1
	add_walls(cell)

	//	While there are walls in the list:
	for wall_list.Len() > 0 {
		rand.Seed(time.Now().UTC().UnixNano())
		rand_wall = int32(rand.Intn(wall_list.Len()))
		ctr = 0
		for e := wall_list.Front(); e != nil; e = e.Next() {
			//	Pick a random wall from the list.
			// 	//  If the cell on the opposite side isn't in the maze yet:
			// 	//	 Make the wall a passage and mark the cell on the opposite side as part of the maze.
			// 	//	 Add the neighboring walls of the cell to the wall list.
			// 	//	If the cell on the opposite side already was in the maze,
			//  //   remove the wall from the list.
			if ctr == rand_wall {
				if DEBUG {
					print(
						" rand_wall: x ", e.Value.(*wall_s).x,
						" y ", e.Value.(*wall_s).y,
						" vert ", e.Value.(*wall_s).vertical,
						" act ", e.Value.(*wall_s).active,
						" e ", e,
						" len(wall_list) ", wall_list.Len(), "\n")
				}
				wall_pos = pos{e.Value.(*wall_s).x, e.Value.(*wall_s).y}

				if e.Value.(*wall_s).vertical == 1 {
					//left side
					if my_maze[wall_pos.x][wall_pos.y] == 0 {
						add_walls(pos{wall_pos.x, wall_pos.y})
						e.Value.(*wall_s).active = 0
						my_maze[wall_pos.x][wall_pos.y] = 1
						// right side
					} else if my_maze[wall_pos.x][wall_pos.y+1] == 0 {
						add_walls(pos{wall_pos.x, wall_pos.y + 1})
						e.Value.(*wall_s).active = 0
						my_maze[wall_pos.x][wall_pos.y+1] = 1
					} else {
						wall_list.Remove(e)
					}
					// horzontal
				} else {
					//top side
					if my_maze[wall_pos.x][wall_pos.y] == 0 {
						add_walls(pos{wall_pos.x, wall_pos.y})
						e.Value.(*wall_s).active = 0
						my_maze[wall_pos.x][wall_pos.y] = 1
						// bottom side
					} else if my_maze[wall_pos.x+1][wall_pos.y] == 0 {
						add_walls(pos{wall_pos.x + 1, wall_pos.y})
						e.Value.(*wall_s).active = 0
						my_maze[wall_pos.x+1][wall_pos.y] = 1
					} else {
						wall_list.Remove(e)
					}
				}
				break
			}
			ctr += 1
		}
	}
}

func make_maze() {
	fill_maze()
	prims()
}

// displacement from box vector origin
var origin_disp = cord{0, -float64(width / 2), -float64(height / 2)}

// size of walls
var my_size = size{1, 1, .25}

type cord struct {
	x, y, z float64
}
type size struct {
	x, y, z float32 // redundent in a sense could cast
}

// creates the box vector
func box(h float32, w float32, d float32) {
	gl.ShadeModel(gl.SMOOTH)
	gl.Normal3d(1.0, 1.0, 1.0)
	//red := [...]float32{0.8, 0.1, 0.0, 1.0}
	//green := [...]float32{0.0, 0.8, 0.2, 1.0}
	//blue := [...]float32{0.2, 0.2, 1.0, 1.0}
	var red float32 = 1.0
	var green float32 = 1.0
	var blue float32 = 1.0

	gl.Materialfv(gl.FRONT, gl.AMBIENT_AND_DIFFUSE, &red)

	var delta float32 = 0
	// left wall
	gl.Begin(gl.POLYGON)
	gl.Color3f(1.0, 0.0, 0.0)
	gl.Vertex3f(delta, delta, delta)
	gl.Vertex3f(delta, h, delta)
	gl.Vertex3f(w, h, delta)
	gl.Vertex3f(w, delta, delta)
	gl.End()

	//right wall
	gl.Begin(gl.POLYGON)
	gl.Color3f(0.0, 0.0, 1.0)
	gl.Vertex3f(delta, delta, d)
	gl.Vertex3f(delta, h, d)
	gl.Vertex3f(w, h, d)
	gl.Vertex3f(w, delta, d)
	gl.End()

	gl.Materialfv(gl.FRONT, gl.AMBIENT_AND_DIFFUSE, &blue)
	//top wall
	gl.Begin(gl.POLYGON)
	gl.Color3f(0.0, 1.0, 0.0)
	gl.Vertex3f(delta, delta, d)
	gl.Vertex3f(w, delta, d)
	gl.Vertex3f(w, delta, delta)
	gl.Vertex3f(delta, delta, delta)
	gl.End()

	//bottom wall
	gl.Begin(gl.POLYGON)
	gl.Color3f(0.5, 0.5, 0.0)
	gl.Vertex3f(delta, h, d)
	gl.Vertex3f(w, h, d)
	gl.Vertex3f(w, h, delta)
	gl.Vertex3f(delta, h, delta)
	gl.End()
	gl.Materialfv(gl.FRONT, gl.AMBIENT_AND_DIFFUSE, &green)

	//forward wall
	gl.Begin(gl.POLYGON)
	gl.Color3f(0.0, 0.0, 1.0)
	gl.Vertex3f(delta, delta, delta)
	gl.Vertex3f(delta, delta, d)
	gl.Vertex3f(delta, h, d)
	gl.Vertex3f(delta, h, delta)
	gl.End()

	// //back wall
	gl.Begin(gl.POLYGON)
	gl.Color3f(0.0, 0.0, 1.0)
	gl.Vertex3f(w, delta, delta)
	gl.Vertex3f(w, delta, d)
	gl.Vertex3f(w, h, d)
	gl.Vertex3f(w, h, delta)
	gl.End()
}

// general draw function
func draw() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT) // CARGOCULT
	gl.PushMatrix()                                     // CARGOCULT
	gl.Rotated(view_rotx, 1.0, 0.0, 0.0)
	gl.Rotated(view_roty, 0.0, 1.0, 0.0)
	gl.Rotated(view_rotz, 0.0, 0.0, 1.0)
	gl.Translated(0.0, 0.0, view_z)

	for i := range boxes {
		gl.PushMatrix() // CARGOCULT
		gl.CallList(boxes[i])
		gl.PopMatrix() // CARGOCULT
	}

	gl.PopMatrix() // CARGOCULT

	sdl.GL_SwapBuffers() // CARGOCULT
}

// meh....
/* new window size or exposure */
func reshape(width int32, height int32) {
	h := float64(height) / float64(width)
	gl.Viewport(0, 0, width, height)
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Frustum(-1.0, 1.0, -h, h, 5.0, 600.0)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Translatef(0.0, 0.0, -100)
}

// list of boxes
// the point of origin
// the translated offset from the origin
// the angles co-ordinate
// the index of the array
// the degree of angular rotation
func wall(b []uint32, o, offset, a_cord cord, i int32, a float64, s size) []uint32 {
	b = append(b, gl.GenLists(1))
	gl.NewList(b[i], gl.COMPILE)
	gl.Translated(o.x+offset.x, o.y+offset.y, o.z+offset.z)
	gl.Rotated(a, a_cord.x, a_cord.y, a_cord.z)
	box(s.x, s.y, s.z)
	gl.EndList()
	return b
}

// void -> void
//  builds the vector array of boxes to be drawn
func init_walls() {
	var ctr = 0         // ctr for box array
	var boff cord       // box offset
	var x_ctr int32 = 0 // x ctr for walls
	var y_ctr int32 = 0 // y ctr for walls

	for _, x := range v_walls {
		for y := range x {
			boff = cord{0, float64(float32(x_ctr) * (my_size.x + my_size.z)), float64(float32(my_size.y) + float32(y_ctr)*(my_size.y+my_size.z))}
			if x[y].active == 1 {
				// draw a vertical wall
				// at position x_ctr y_ctt
				boxes = wall(boxes, origin_disp, boff, cord{0, 0, 0}, int32(ctr), 0, my_size) // top
				ctr++
			}
			y_ctr++
		}
		x_ctr++
		y_ctr = 0
	}

	for _, x := range h_walls {
		for y := range x {
			boff = cord{0, float64(float32(x_ctr)*(my_size.x+my_size.z) - float32(height-1)*(my_size.x+my_size.z)), float64(float32(y_ctr) * (my_size.y + my_size.z))}
			if x[y].active == 1 {
				// draw horizontal wall
				// at position x_ctr y_ctr
				boxes = wall(boxes, origin_disp, boff, cord{1, 0, 0}, int32(ctr), 90, my_size) // left
				ctr++
			}
			y_ctr++
		}
		x_ctr++
		y_ctr = 0
	}
	x_ctr = 0
	y_ctr = 0

	// draw border

}

// lets get things rolling
func init_() {
	// all for lighting effects
	//pos := []float32{0.0, 0.0, 100.0, 0.0}
	var pos float32 = 1.0
	gl.Lightfv(gl.LIGHT0, gl.POSITION, &pos)
	gl.Enable(gl.LIGHTING)
	gl.Enable(gl.LIGHT0)
	// below is necessary
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.NORMALIZE)

	if DEBUG {
		tests()
		init_walls()
		return
	} else {
		make_maze()
		init_walls()
	}
}

// returns if we are done or not
func key_handler() bool {
	var keys []uint8 = sdl.GetKeyState()

	if keys[sdl.K_i] != 0 {
		view_z += 1
	}
	if keys[sdl.K_o] != 0 {
		view_z -= 1
	}
	if keys[sdl.K_ESCAPE] != 0 {
		return true
	}
	if keys[sdl.K_UP] != 0 {
		view_rotx += 1
	}
	if keys[sdl.K_DOWN] != 0 {
		view_rotx -= 1.0
	}
	if keys[sdl.K_LEFT] != 0 {
		view_roty += 1.0
	}
	if keys[sdl.K_RIGHT] != 0 {
		view_roty -= 1.0
	}
	if keys[sdl.K_z] != 0 {
		if (sdl.GetModState() & sdl.KMOD_RSHIFT) != 0 {
			view_rotz -= 1.0
		} else {
			view_rotz += 1.0
		}
	}
	if keys[sdl.K_w] != 0 {
		gl.Translatef(0.0, 0.0, 1)
	}
	if keys[sdl.K_s] != 0 {
		gl.Translatef(0.0, 0.0, -1)
	}
	if keys[sdl.K_a] != 0 {
		gl.Translatef(-1, 0, 0)
	}
	if keys[sdl.K_d] != 0 {
		gl.Translatef(1, 0, 0)
	}
	if keys[sdl.K_q] != 0 {
		gl.Translatef(0, 1, 0)
	}
	if keys[sdl.K_e] != 0 {
		gl.Translatef(0, -1, 0)
	}
	return false
}

func main() {
	var done bool

	sdl.Init(sdl.INIT_VIDEO)

	var screen = sdl.SetVideoMode(300, 300, 18, sdl.OPENGL|sdl.RESIZABLE)

	if screen == nil {
		sdl.Quit()
		panic("Couldn't set 300x300 GL video mode: " + sdl.GetError() + "\n")
	}
	gl.Init()
	//if gl.Init() != nil {
	//panic("gl error")
	//}
	init_()
	reshape(int32(screen.W), int32(screen.H))
	done = false
	for !done {
		for e := sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
			switch e.(type) {
			case *sdl.ResizeEvent:
				re := e.(*sdl.ResizeEvent)
				screen = sdl.SetVideoMode(int(re.W), int(re.H), 16,
					sdl.OPENGL|sdl.RESIZABLE)
				if screen != nil {
					reshape(int32(screen.W), int32(screen.H))
				} else {
					panic("we couldn't set the new video mode??")
				}
				break
			case *sdl.QuitEvent:
				done = true
				break
			}
		}
		done = key_handler()
		draw()
	}
	sdl.Quit()
	return
}

func tests() {
	fill_maze()
	var w int32 = width - 1
	var h int32 = height - 1

	var tl = pos{0, 0}
	var bl = pos{h, 0}
	var tr = pos{0, w}
	var br = pos{h, w}

	var test_arr [4]pos
	test_arr[0] = tl
	test_arr[1] = bl
	test_arr[2] = tr
	test_arr[3] = br

	h_walls[1][1].active = 0
	h_walls[2][2].active = 0
	h_walls[3][3].active = 0

	var top_left = pos{0, 0}
	var bot_rght = pos{h, w}
	var top_rght = pos{0, w}
	var bot_left = pos{h, 0}
	print("top left test \n")
	add_walls(top_left)
	// test wall_list length
	print("bottom right test\n")
	add_walls(bot_rght)
	print("top right test\n")
	add_walls(top_rght)
	print("bottom left test\n")
	add_walls(bot_left)
	print("\n")
}
