package core

import (
	"time"

	"fyne.io/fyne/v2"
)

const EPSILON = 1e-9

const STRING_EMPTY = ""
const STRING_PLUS = "+"
const STRING_MINUS = "-"
const STRING_PERCENTAGE = "%"
const STRING_PERCENTAGE_DIVIDE = "/100"
const STRING_DOLLAR = "$"
const STRING_PIPE = "|"
const STRING_DOUBLE_EQUAL = "=="
const STRING_EQUAL = "="
const STRING_NOT_EQUAL = "!="
const STRING_LESS = "<"
const STRING_GREATER = ">"
const STRING_LESS_EQUAL = "<="
const STRING_GREATER_EQUAL = ">="
const STRING_ELLIPISIS = "..."

const FMT_SHORT_TRILLION_DOLLAR = "$%.2fT"
const FMT_SHORT_BILLION_DOLLAR = "$%.2fB"
const FMT_SHORT_MILLION_DOLLAR = "$%.2fM"
const FMT_SHORT_THOUSAND_DOLLAR = "$%.2fK"
const FMT_SHORT_DOLLAR = "$%.2f"

const NETWORKING_SUCCESS = 0
const NETWORKING_ERROR_CONNECTION = -1
const NETWORKING_URL_ERROR = -2
const NETWORKING_DATA_IN_CACHE = -3
const NETWORKING_UNAUTHORIZED = -4
const NETWORKING_BAD_DATA_RECEIVED = -5
const NETWORKING_BAD_CONFIG = -6
const NETWORKING_BAD_PAYLOAD = -7
const NETWORKING_FAILED_CREATE_FILE = -8

const STATE_LOADED = 0
const STATE_LOADING = -1
const STATE_ERROR = -2
const STATE_FETCHING_NEW = -3
const STATE_BAD_CONFIG = -4

const VALUE_INCREASE = 1
const VALUE_NO_CHANGE = 0
const VALUE_DECREASE = -1

const NO_SNAPSHOT = -1
const HAVE_SNAPSHOT = 0
const SNAPSHOT_DELETED = 1
const SNAPSHOT_DELETE_FAILED = 2
const MINIMUM_SNAPSHOT_SAVE_INTERVAL = 30 * time.Minute

const STATUS_SUCCESS = 0
const STATUS_NETWORK_ERROR = 1
const STATUS_CONFIG_ERROR = 2
const STATUS_BAD_DATA_RECEIVED = 3
const STATUS_CANCELLED = 4

const ColorNameError fyne.ThemeColorName = "error"
const ColorNameTransparent fyne.ThemeColorName = "transparent"
const ColorNameRed fyne.ThemeColorName = "red"
const ColorNameDarkRed fyne.ThemeColorName = "darkRed"
const ColorNameGreen fyne.ThemeColorName = "green"
const ColorNameDarkGreen fyne.ThemeColorName = "darkGreen"
const ColorNameBlue fyne.ThemeColorName = "blue"
const ColorNameLightBlue fyne.ThemeColorName = "lightBlue"
const ColorNameLightPurple fyne.ThemeColorName = "lightPurple"
const ColorNameLightOrange fyne.ThemeColorName = "lightOrange"
const ColorNameOrange fyne.ThemeColorName = "orange"
const ColorNameYellow fyne.ThemeColorName = "yellow"
const ColorNameTeal fyne.ThemeColorName = "teal"
const ColorNameDarkGrey fyne.ThemeColorName = "darkGrey"
const ColorNamePanelBG fyne.ThemeColorName = "panelBG"
const ColorNamePanelPlaceholder fyne.ThemeColorName = "panelPlaceholder"
const ColorNameTickerBG fyne.ThemeColorName = "tickerBG"

const SizeLayoutPadding fyne.ThemeSizeName = "layoutPadding"
const SizePanelBorderRadius fyne.ThemeSizeName = "panelBorderRadius"
const SizePanelTitle fyne.ThemeSizeName = "panelTitle"
const SizePanelSubTitle fyne.ThemeSizeName = "panelSubTitle"
const SizePanelBottomText fyne.ThemeSizeName = "panelBottomText"
const SizePanelContent fyne.ThemeSizeName = "panelContent"
const SizePanelTitleSmall fyne.ThemeSizeName = "panelTitleSmall"
const SizePanelSubTitleSmall fyne.ThemeSizeName = "panelSubTitleSmall"
const SizePanelBottomTextSmall fyne.ThemeSizeName = "panelBottomTextSmall"
const SizePanelContentSmall fyne.ThemeSizeName = "panelContentSmall"
const SizePanelWidth fyne.ThemeSizeName = "panelWidth"
const SizePanelHeight fyne.ThemeSizeName = "panelHeight"
const SizeActionBtnWidth fyne.ThemeSizeName = "actionBtnWidth"
const SizeActionBtnGap fyne.ThemeSizeName = "actionBtnGap"
const SizeTickerBorderRadius fyne.ThemeSizeName = "tickerBorderRadius"
const SizeTickerWidth fyne.ThemeSizeName = "tickerWidth"
const SizeTickerHeight fyne.ThemeSizeName = "tickerHeight"
const SizeTickerTitle fyne.ThemeSizeName = "tickerTitle"
const SizeTickerContent fyne.ThemeSizeName = "tickerContent"
const SizeNotificationText fyne.ThemeSizeName = "notificationText"
const SizeCompletionText fyne.ThemeSizeName = "completionText"
const SizePaddingPanelLeft fyne.ThemeSizeName = "paddingPanelLeft"
const SizePaddingPanelTop fyne.ThemeSizeName = "paddingPanelTop"
const SizePaddingPanelRight fyne.ThemeSizeName = "paddingPanelRight"
const SizePaddingPanelBottom fyne.ThemeSizeName = "paddingPanelBottom"

const SizePaddingTickerLeft fyne.ThemeSizeName = "paddingTickerLeft"
const SizePaddingTickerTop fyne.ThemeSizeName = "paddingTickerTop"
const SizePaddingTickerRight fyne.ThemeSizeName = "paddingTickerRight"
const SizePaddingTickerBottom fyne.ThemeSizeName = "paddingTickerBottom"
