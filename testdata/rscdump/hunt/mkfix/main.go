package main
import (
  "github.com/gen0cide/westworld/internal/rscdump"
)
func mk(path string, id int) {
  const size=16; n:=size*size
  d:=&rscdump.Dump{Schema:rscdump.SchemaID,Level:rscdump.LevelL1,Source:rscdump.SourceHandAuthored,
    Camera:rscdump.Camera{Yaw:400,Distance:1100,ScreenW:512,ScreenH:334,ViewDist:9,ClipNear:5,ClipFar:7000},
    Window:rscdump.Window{BaseX:0,BaseY:0,Plane:0,Size:size},
    Terrain:&rscdump.Terrain{Size:size,Elevation:make([]byte,n),GroundColour:make([]byte,n),TerrainSeed:0},
    Scenery:[]rscdump.Scenery{{X:8,Y:8,ID:id,Dir:0}}}
  if err:=d.Save(path); err!=nil { panic(err) }
}
func main(){
  mk("/home/free/code/rsc-hacking/westworld/testdata/rscdump/hunt/scenery_windmill_sails_id74.json",74)
  mk("/home/free/code/rsc-hacking/westworld/testdata/rscdump/hunt/scenery_windmill_sails_id900.json",900)
}
