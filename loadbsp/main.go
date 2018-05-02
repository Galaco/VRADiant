package loadbsp

import (
	"os"
	"strconv"
	"log"
	"strings"
	"github.com/galaco/vrad/common"
	"github.com/galaco/vrad/cmd"
	"github.com/galaco/vrad/vmath/vector"
	"github.com/galaco/vrad/vmath/matrix"
	"github.com/galaco/bsp"
	"github.com/galaco/bsp/lumps"
	"github.com/galaco/bsp/primitives/mapflags"
	"github.com/galaco/bsp/primitives/face"
	"github.com/galaco/vmf"
	"github.com/galaco/bsp/primitives/model"
	"github.com/galaco/vrad/loadbsp/brush"
)

func Main(args *cmd.Args, transfered interface{}) (interface{}, error) {
	if args.LowPriority == true {
		// Go doesnt really support this...
	}

	if args.DumpPatches == true {
		//InitDumpPatchesFiles()
	}

	// Here we should prep the log file...

	if args.Lights != "" {
		// Load lights.rad!!!
	}

	log.Printf("Loading %s\n", args.Filename)
	file,err := loadBSP(args.Filename)
	if err != nil {
		return nil,err
	}
	common.GLOBALSET_BSP(file)

	//g_pFullFileSystem->AddSearchPath(source, "GAME", PATH_ADD_TO_HEAD);
	//g_pFullFileSystem->AddSearchPath(source, "MOD", PATH_ADD_TO_HEAD);
	mapFlagsLump := (*common.GLOBALGET_BSP().GetLump(bsp.LUMP_MAP_FLAGS).GetContents()).(lumps.MapFlags)
	mapFlags := mapFlagsLump.GetData().(*mapflags.MapFlags)
	if args.StaticPropLighting {
		mask := 0x00000002
		if args.HDR {
			mask = 0x00000001
		}
		mapFlags.LevelFlags |= uint32(mask)
	} else {
		// @TODO This *could* be incorrect
		mapFlags.LevelFlags &^= uint32(0x00000002 | 0x00000001)
	}

	// Determine face target
	var targetFaces []face.Face
	var numFaces = 0
	if args.HDR == true {
		targetFaces = (*common.GLOBALGET_BSP().GetLump(bsp.LUMP_FACES_HDR).GetContents()).(lumps.FaceHDR).GetData().([]face.Face)
		if len(targetFaces) == 0 {
			numFaces = len((*common.GLOBALGET_BSP().GetLump(bsp.LUMP_FACES).GetContents()).(lumps.Face).GetData().([]face.Face))
		}
	} else {
		targetFaces = *(*common.GLOBALGET_BSP().GetLump(bsp.LUMP_FACES).GetContents()).(lumps.Face).GetData().(*[]face.Face)
	}

	entData := (*common.GLOBALGET_BSP().GetLump(bsp.LUMP_ENTITIES).GetContents()).(lumps.EntData)
	entImportAsVmfBlock,err := parseEntities(&entData)
	entities := entImportAsVmfBlock.Unclassified

	ExtractBrushEntityShadowCasters(&entities)

	//StaticPropMgr()->Init();
	//StaticDispMgr()->Init();

	/**if !visdatasize {
		log.Printf("No vis information, direct lighting only.\n")
		numbounce = 0;
		ambient[0] = ambient[1] = ambient[2] = 0.1
		dvis->numclusters = CountClusters()
	}

	//
	// patches and referencing data (ensure capacity)
	//
	// TODO: change the maxes to the amount from the bsp!!
	//
	//	g_Patches.EnsureCapacity( MAX_PATCHES );

	g_FacePatches.SetSize( MAX_MAP_FACES );
	faceParents.SetSize( MAX_MAP_FACES );
	clusterChildren.SetSize( MAX_MAP_CLUSTERS );

	int ndx;
	for ( ndx = 0; ndx < MAX_MAP_FACES; ndx++ )
	{
	g_FacePatches[ndx] = g_FacePatches.InvalidIndex();
	faceParents[ndx] = faceParents.InvalidIndex();
	}

	for ( ndx = 0; ndx < MAX_MAP_CLUSTERS; ndx++ )
	{
	clusterChildren[ndx] = clusterChildren.InvalidIndex();
	}**/


	return numFaces,nil
}

func loadBSP(filename string) (*bsp.Bsp,error){
	file,err := os.Open(filename)
	if err != nil {
		return nil,err
	}
	reader := bsp.NewReader(file)
	return reader.Read()
}

func parseEntities(data *lumps.EntData) (vmf.Vmf,error) {
	stringReader := strings.NewReader(*data.GetData().(*string))
	reader := vmf.NewReader(stringReader)

	return reader.Read()
}


// Some brush entities can cast shadows.
// We need to make a note of them
func ExtractBrushEntityShadowCasters(entities *vmf.Node) {
	models := (*common.GLOBALGET_BSP().GetLump(bsp.LUMP_MODELS).GetContents()).(lumps.Model).GetData().(*[]model.Model)
	for _,iEntity := range *entities.GetAllValues() {
		entity := iEntity.(vmf.Node)
		if entity.HasProperty("vrad_brush_cast_shadows") == true {
			splOrigin := strings.Split(entity.GetProperty("origin"), " ")
			splAngles := strings.Split(entity.GetProperty("angles"), " ")
			origin := vector.Vec3{}
			for i,sf := range splOrigin {
				f,_ := strconv.ParseFloat(sf, 32)
				origin[i] = float32(f)
			}
			angles := vector.Vec3{}
			for i,sf := range splAngles {
				f,_ := strconv.ParseFloat(sf, 32)
				angles[i] = float32(f)
			}

			xform := matrix.Mat4{}
			xform.SetupMatrixOrgAngles( &origin, &angles )
			// Adds to raytrace environment
			addBrushes(brushmodelForEntity(&entity, models), xform)
		}
	}
}

func brushmodelForEntity(entity *vmf.Node, models *[]model.Model) *model.Model {
	modelName := entity.GetProperty("model")
	if len(modelName) > 1 {
		modelIndex,_ := strconv.ParseInt(modelName, 8, 32)
		modelIndex += 1
		if modelIndex > 0 && int(modelIndex) < len(*models) {
			return &(*models)[modelIndex]
		}
	}

	return nil
}

func addBrushes(model *model.Model, xform matrix.Mat4) {
	if model != nil {
		brushList := []int{}

		brush.GetBrushRecursive(int(model.HeadNode), &brushList)
		for i := 0; i < len(brushList); i++ {
			//ndxBrush := brushList[i]
			//AddBrushToRaytraceEnvironment( &dbrushes[ndxBrush], xform )
		}
	}
}

/**
dmodel_t *BrushmodelForEntity( entity_t *pEntity )
{
	const char *pModelname = ValueForKey( pEntity, "model" );
	if ( Q_strlen(pModelname) > 1 )
	{
		int modelIndex = atol( pModelname + 1 );
		if ( modelIndex > 0 && modelIndex < nummodels )
		{
			return &dmodels[modelIndex];
		}
	}
	return NULL;
}
 */


/**
ThreadSetDefault ();

	g_flStartTime = Plat_FloatTime();

	if( g_bLowPriority )
	{
		SetLowPriority();
	}

	strcpy( level_name, source );

	// This must come after InitFileSystem because the file system pointer might change.
	if ( g_bDumpPatches )
		InitDumpPatchesFiles();

	// This part is just for VMPI. VMPI's file system needs the basedir in front of all filenames,
	// so we prepend qdir here.
	strcpy( source, ExpandPath( source ) );

	if ( !g_bUseMPI )
	{
		// Setup the logfile.
		char logFile[512];
		_snprintf( logFile, sizeof(logFile), "%s.log", source );
		SetSpewFunctionLogFile( logFile );
	}

	LoadPhysicsDLL();

	// Set the required global lights filename and try looking in qproject
	strcpy( global_lights, "lights.rad" );
	if ( !g_pFileSystem->FileExists( global_lights ) )
	{
		// Otherwise, try looking in the BIN directory from which we were run from
		Msg( "Could not find lights.rad in %s.\nTrying VRAD BIN directory instead...\n",
			    global_lights );
		GetModuleFileName( NULL, global_lights, sizeof( global_lights ) );
		Q_ExtractFilePath( global_lights, global_lights, sizeof( global_lights ) );
		strcat( global_lights, "lights.rad" );
	}

	// Set the optional level specific lights filename
	strcpy( level_lights, source );

	Q_DefaultExtension( level_lights, ".rad", sizeof( level_lights ) );
	if ( !g_pFileSystem->FileExists( level_lights ) )
		*level_lights = 0;

	ReadLightFile(global_lights);							// Required
	if ( *designer_lights ) ReadLightFile(designer_lights);	// Command-line
	if ( *level_lights )	ReadLightFile(level_lights);	// Optional & implied

	strcpy(incrementfile, source);
	Q_DefaultExtension(incrementfile, ".r0", sizeof(incrementfile));
	Q_DefaultExtension(source, ".bsp", sizeof( source ));

	Msg( "Loading %s\n", source );
	VMPI_SetCurrentStage( "LoadBSPFile" );
	LoadBSPFile (source);

	// Add this bsp to our search path so embedded resources can be found
	if ( g_bUseMPI && g_bMPIMaster )
	{
		// MPI Master, MPI workers don't need to do anything
		g_pOriginalPassThruFileSystem->AddSearchPath(source, "GAME", PATH_ADD_TO_HEAD);
		g_pOriginalPassThruFileSystem->AddSearchPath(source, "MOD", PATH_ADD_TO_HEAD);
	}
	else if ( !g_bUseMPI )
	{
		// Non-MPI
		g_pFullFileSystem->AddSearchPath(source, "GAME", PATH_ADD_TO_HEAD);
		g_pFullFileSystem->AddSearchPath(source, "MOD", PATH_ADD_TO_HEAD);
	}

	// now, set whether or not static prop lighting is present
	if (g_bStaticPropLighting)
		g_LevelFlags |= g_bHDR? LVLFLAGS_BAKED_STATIC_PROP_LIGHTING_HDR : LVLFLAGS_BAKED_STATIC_PROP_LIGHTING_NONHDR;
	else
	{
		g_LevelFlags &= ~( LVLFLAGS_BAKED_STATIC_PROP_LIGHTING_HDR | LVLFLAGS_BAKED_STATIC_PROP_LIGHTING_NONHDR );
	}

	// now, we need to set our face ptr depending upon hdr, and if hdr, init it
	if (g_bHDR)
	{
		g_pFaces = dfaces_hdr;
		if (numfaces_hdr==0)
		{
			numfaces_hdr = numfaces;
			memcpy( dfaces_hdr, dfaces, numfaces*sizeof(dfaces[0]) );
		}
	}
	else
	{
		g_pFaces = dfaces;
	}


	ParseEntities ();
	ExtractBrushEntityShadowCasters();

	StaticPropMgr()->Init();
	StaticDispMgr()->Init();

	if (!visdatasize)
	{
		Msg("No vis information, direct lighting only.\n");
		numbounce = 0;
		ambient[0] = ambient[1] = ambient[2] = 0.1f;
		dvis->numclusters = CountClusters();
	}

	//
	// patches and referencing data (ensure capacity)
	//
	// TODO: change the maxes to the amount from the bsp!!
	//
//	g_Patches.EnsureCapacity( MAX_PATCHES );

	g_FacePatches.SetSize( MAX_MAP_FACES );
	faceParents.SetSize( MAX_MAP_FACES );
	clusterChildren.SetSize( MAX_MAP_CLUSTERS );

	int ndx;
	for ( ndx = 0; ndx < MAX_MAP_FACES; ndx++ )
	{
		g_FacePatches[ndx] = g_FacePatches.InvalidIndex();
		faceParents[ndx] = faceParents.InvalidIndex();
	}

	for ( ndx = 0; ndx < MAX_MAP_CLUSTERS; ndx++ )
	{
		clusterChildren[ndx] = clusterChildren.InvalidIndex();
	}

	// Setup ray tracer
	AddBrushesForRayTrace();
	StaticDispMgr()->AddPolysForRayTrace();
	StaticPropMgr()->AddPolysForRayTrace();

	// Dump raytracer for glview
	if ( g_bDumpRtEnv )
		WriteRTEnv("trace.txt");

	// Build acceleration structure
	printf ( "Setting up ray-trace acceleration structure... ");
	float start = Plat_FloatTime();
	g_RtEnv.SetupAccelerationStructure();
	float end = Plat_FloatTime();
	printf ( "Done (%.2f seconds)\n", end-start );

#if 0  // To test only k-d build
	exit(0);
#endif

	RadWorld_Start();

	// Setup incremental lighting.
	if( g_pIncremental )
	{
		if( !g_pIncremental->Init( source, incrementfile ) )
		{
			Error( "Unable to load incremental lighting file in %s.\n", incrementfile );
			return;
		}
	}
 */