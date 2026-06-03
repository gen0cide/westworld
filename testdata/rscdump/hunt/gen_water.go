//go:build ignore
package main
import "github.com/gen0cide/westworld/internal/rscdump"
const ( Size=24; BaseX=200; BaseY=200; Elev=40 )
func idx(x,y int) int { return x*Size+y }
func cam() rscdump.Camera { return rscdump.Camera{X:0,Y:-384,Z:0,Pitch:912,Yaw:512,Roll:0,Distance:1800,ViewDist:9,ClipNear:5,ClipFar:7000,ScreenW:512,ScreenH:334} }
func base(ov []byte) *rscdump.Dump {
  n:=Size*Size; elev:=make([]byte,n); g:=make([]byte,n)
  for i:=range elev { elev[i]=Elev; g[i]=70 }
  return &rscdump.Dump{Schema:rscdump.SchemaID,Level:rscdump.LevelL1,Source:rscdump.SourceHandAuthored,Camera:cam(),
    Window:rscdump.Window{BaseX:BaseX,BaseY:BaseY,Plane:0,Size:Size},
    Terrain:&rscdump.Terrain{Size:Size,Elevation:elev,GroundColour:g,Overlay:ov,TerrainSeed:0},
    Self:&rscdump.Self{X:BaseX+Size/2,Y:BaseY+Size/2,NoSelf:true}}
}
func main(){
  ov2:=make([]byte,Size*Size); for x:=8;x<=15;x++{for y:=8;y<=15;y++{ov2[idx(x,y)]=2}}
  base(ov2).Save("/tmp/terrain_water_ov2.json")
}
