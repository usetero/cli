package styles

// Tailwind CSS v4 color palette (converted from OKLCH to hex)
// https://tailwindcss.com/docs/customizing-colors

// Shade represents a Tailwind color shade (50, 100, 200, ... 900, 950)
type Shade int

const (
	S50  Shade = 50
	S100 Shade = 100
	S200 Shade = 200
	S300 Shade = 300
	S400 Shade = 400
	S500 Shade = 500
	S600 Shade = 600
	S700 Shade = 700
	S800 Shade = 800
	S900 Shade = 900
	S950 Shade = 950
)

// ColorFamily maps shades to hex colors
type ColorFamily map[Shade]string

// Color families - Tailwind colors organized by shade
var (
	SlateFamily = ColorFamily{
		S50: "#F8FAFC", S100: "#F1F5F9", S200: "#E2E8F0", S300: "#CBD5E1",
		S400: "#94A3B8", S500: "#64748B", S600: "#475569", S700: "#334155",
		S800: "#1E293B", S900: "#0F172A", S950: "#020617",
	}
	GrayFamily = ColorFamily{
		S50: "#F9FAFB", S100: "#F3F4F6", S200: "#E5E7EB", S300: "#D1D5DB",
		S400: "#9CA3AF", S500: "#6B7280", S600: "#4B5563", S700: "#374151",
		S800: "#1F2937", S900: "#111827", S950: "#030712",
	}
	ZincFamily = ColorFamily{
		S50: "#FAFAFA", S100: "#F4F4F5", S200: "#E4E4E7", S300: "#D4D4D8",
		S400: "#A1A1AA", S500: "#71717A", S600: "#52525B", S700: "#3F3F46",
		S800: "#27272A", S900: "#18181B", S950: "#09090B",
	}
	NeutralFamily = ColorFamily{
		S50: "#FAFAFA", S100: "#F5F5F5", S200: "#E5E5E5", S300: "#D4D4D4",
		S400: "#A3A3A3", S500: "#737373", S600: "#525252", S700: "#404040",
		S800: "#262626", S900: "#171717", S950: "#0A0A0A",
	}
	StoneFamily = ColorFamily{
		S50: "#FAFAF9", S100: "#F5F5F4", S200: "#E7E5E4", S300: "#D6D3D1",
		S400: "#A8A29E", S500: "#78716C", S600: "#57534E", S700: "#44403C",
		S800: "#292524", S900: "#1C1917", S950: "#0C0A09",
	}
	RedFamily = ColorFamily{
		S50: "#FEF2F2", S100: "#FEE2E2", S200: "#FECACA", S300: "#FCA5A5",
		S400: "#F87171", S500: "#EF4444", S600: "#DC2626", S700: "#B91C1C",
		S800: "#991B1B", S900: "#7F1D1D", S950: "#450A0A",
	}
	OrangeFamily = ColorFamily{
		S50: "#FFF7ED", S100: "#FFEDD5", S200: "#FED7AA", S300: "#FDBA74",
		S400: "#FB923C", S500: "#F97316", S600: "#EA580C", S700: "#C2410C",
		S800: "#9A3412", S900: "#7C2D12", S950: "#431407",
	}
	AmberFamily = ColorFamily{
		S50: "#FFFBEB", S100: "#FEF3C7", S200: "#FDE68A", S300: "#FCD34D",
		S400: "#FBBF24", S500: "#F59E0B", S600: "#D97706", S700: "#B45309",
		S800: "#92400E", S900: "#78350F", S950: "#451A03",
	}
	YellowFamily = ColorFamily{
		S50: "#FEFCE8", S100: "#FEF9C3", S200: "#FEF08A", S300: "#FDE047",
		S400: "#FACC15", S500: "#EAB308", S600: "#CA8A04", S700: "#A16207",
		S800: "#854D0E", S900: "#713F12", S950: "#422006",
	}
	LimeFamily = ColorFamily{
		S50: "#F7FEE7", S100: "#ECFCCB", S200: "#D9F99D", S300: "#BEF264",
		S400: "#A3E635", S500: "#84CC16", S600: "#65A30D", S700: "#4D7C0F",
		S800: "#3F6212", S900: "#365314", S950: "#1A2E05",
	}
	GreenFamily = ColorFamily{
		S50: "#F0FDF4", S100: "#DCFCE7", S200: "#BBF7D0", S300: "#86EFAC",
		S400: "#4ADE80", S500: "#22C55E", S600: "#16A34A", S700: "#15803D",
		S800: "#166534", S900: "#14532D", S950: "#052E16",
	}
	EmeraldFamily = ColorFamily{
		S50: "#ECFDF5", S100: "#D1FAE5", S200: "#A7F3D0", S300: "#6EE7B7",
		S400: "#34D399", S500: "#10B981", S600: "#059669", S700: "#047857",
		S800: "#065F46", S900: "#064E3B", S950: "#022C22",
	}
	TealFamily = ColorFamily{
		S50: "#F0FDFA", S100: "#CCFBF1", S200: "#99F6E4", S300: "#5EEAD4",
		S400: "#2DD4BF", S500: "#14B8A6", S600: "#0D9488", S700: "#0F766E",
		S800: "#115E59", S900: "#134E4A", S950: "#042F2E",
	}
	CyanFamily = ColorFamily{
		S50: "#ECFEFF", S100: "#CFFAFE", S200: "#A5F3FC", S300: "#67E8F9",
		S400: "#22D3EE", S500: "#06B6D4", S600: "#0891B2", S700: "#0E7490",
		S800: "#155E75", S900: "#164E63", S950: "#083344",
	}
	SkyFamily = ColorFamily{
		S50: "#F0F9FF", S100: "#E0F2FE", S200: "#BAE6FD", S300: "#7DD3FC",
		S400: "#38BDF8", S500: "#0EA5E9", S600: "#0284C7", S700: "#0369A1",
		S800: "#075985", S900: "#0C4A6E", S950: "#082F49",
	}
	BlueFamily = ColorFamily{
		S50: "#EFF6FF", S100: "#DBEAFE", S200: "#BFDBFE", S300: "#93C5FD",
		S400: "#60A5FA", S500: "#3B82F6", S600: "#2563EB", S700: "#1D4ED8",
		S800: "#1E40AF", S900: "#1E3A8A", S950: "#172554",
	}
	IndigoFamily = ColorFamily{
		S50: "#EEF2FF", S100: "#E0E7FF", S200: "#C7D2FE", S300: "#A5B4FC",
		S400: "#818CF8", S500: "#6366F1", S600: "#4F46E5", S700: "#4338CA",
		S800: "#3730A3", S900: "#312E81", S950: "#1E1B4B",
	}
	VioletFamily = ColorFamily{
		S50: "#F5F3FF", S100: "#EDE9FE", S200: "#DDD6FE", S300: "#C4B5FD",
		S400: "#A78BFA", S500: "#8B5CF6", S600: "#7C3AED", S700: "#6D28D9",
		S800: "#5B21B6", S900: "#4C1D95", S950: "#2E1065",
	}
	PurpleFamily = ColorFamily{
		S50: "#FAF5FF", S100: "#F3E8FF", S200: "#E9D5FF", S300: "#D8B4FE",
		S400: "#C084FC", S500: "#A855F7", S600: "#9333EA", S700: "#7E22CE",
		S800: "#6B21A8", S900: "#581C87", S950: "#3B0764",
	}
	FuchsiaFamily = ColorFamily{
		S50: "#FDF4FF", S100: "#FAE8FF", S200: "#F5D0FE", S300: "#F0ABFC",
		S400: "#E879F9", S500: "#D946EF", S600: "#C026D3", S700: "#A21CAF",
		S800: "#86198F", S900: "#701A75", S950: "#4A044E",
	}
	PinkFamily = ColorFamily{
		S50: "#FDF2F8", S100: "#FCE7F3", S200: "#FBCFE8", S300: "#F9A8D4",
		S400: "#F472B6", S500: "#EC4899", S600: "#DB2777", S700: "#BE185D",
		S800: "#9D174D", S900: "#831843", S950: "#500724",
	}
	RoseFamily = ColorFamily{
		S50: "#FFF1F2", S100: "#FFE4E6", S200: "#FECDD3", S300: "#FDA4AF",
		S400: "#FB7185", S500: "#F43F5E", S600: "#E11D48", S700: "#BE123C",
		S800: "#9F1239", S900: "#881337", S950: "#4C0519",
	}
)

// White is a special case - not part of a family
const White = "#FFFFFF"
