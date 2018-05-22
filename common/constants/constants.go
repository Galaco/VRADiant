package constants

const MAX_COORD_INTEGER = 16384

// We can have larger lightmaps on displacements
const MAX_DISP_LIGHTMAP_DIM_WITHOUT_BORDER	= 125
const MAX_DISP_LIGHTMAP_DIM_INCLUDING_BORDER =128


// This is the actual max.. (change if you change the brush lightmap dim or disp lightmap dim
const MAX_LIGHTMAP_DIM_WITHOUT_BORDER	= MAX_DISP_LIGHTMAP_DIM_WITHOUT_BORDER
const MAX_LIGHTMAP_DIM_INCLUDING_BORDER	= MAX_DISP_LIGHTMAP_DIM_INCLUDING_BORDER


const CONSTRUCTS_INVALID_INDEX = -1
const MAX_POINTS_ON_WINDING = 64
const NUM_BUMP_VECTS = 3

const ANGLE_UP = -1
const ANGLE_DOWN = -2

const PITCH = 0
const YAW = 1
const ROLL = 2


const TEMPCONST_NUM_THREADS = 1