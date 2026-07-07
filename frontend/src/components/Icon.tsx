import { cn } from '@/lib/utils';
import { useTheme } from '../lib/theme';
import type { ComponentType } from 'react';

// Namespace imports — forces entire module into bundle
import * as Lu from 'react-icons/lu';
import * as Fa6 from 'react-icons/fa6';
import * as Hi2 from 'react-icons/hi2';
import * as Md from 'react-icons/md';
import * as Ri from 'react-icons/ri';

// Export the namespace modules so Vite knows they're used
export const _usedIcons = [Lu, Fa6, Hi2, Md, Ri];

const setConfig: Record<string, { pkg: any; names: Record<string, string> }> = {
  modern: {
    pkg: Lu,
    names: {
      Edit: 'LuPencil',
      Check: 'LuCheck',
      Trash2: 'LuTrash2',
      Plus: 'LuPlus',
      Save: 'LuSave',
      X: 'LuX',
      Search: 'LuSearch',
      Maximize2: 'LuMaximize2',
      Minimize2: 'LuMinimize2',
      FileText: 'LuFileText',
      Moon: 'LuMoon',
      Sun: 'LuSun',
      Settings: 'LuSettings',
      Folder: 'LuFolder',
      ChevronDown: 'LuChevronDown',
      ChevronRight: 'LuChevronRight',
      User: 'LuUser',
      Home: 'LuHouse',
      Brain: 'LuBrain',
      Wrench: 'LuWrench',
      Layers: 'LuLayers',
      Star: 'LuStar',
      Cpu: 'LuCpu',
      Code: 'LuCode',
      Zap: 'LuZap',
      MessageSquare: 'LuMessageSquare',
      Sparkles: 'LuSparkles',
      Eye: 'LuEye',
      Network: 'LuNetwork',
      Bot: 'LuBot',
    },
  },
  classic: {
    pkg: Fa6,
    names: {
      Edit: 'FaPen',
      Check: 'FaCheck',
      Trash2: 'FaTrashCan',
      Plus: 'FaPlus',
      Save: 'FaFloppyDisk',
      X: 'FaXmark',
      Search: 'FaMagnifyingGlass',
      Maximize2: 'FaExpand',
      Minimize2: 'FaCompress',
      FileText: 'FaFileLines',
      Moon: 'FaMoon',
      Sun: 'FaSun',
      Settings: 'FaGear',
      Folder: 'FaFolder',
      ChevronDown: 'FaChevronDown',
      ChevronRight: 'FaChevronRight',
      User: 'FaUser',
      Home: 'FaHouse',
      Brain: 'FaBrain',
      Wrench: 'FaWrench',
      Star: 'FaStar',
      Code: 'FaCode',
      Zap: 'FaBolt',
      MessageSquare: 'FaMessage',
      Sparkles: 'FaStar',
      Eye: 'FaEye',
      Network: 'FaNetworkWired',
    },
  },
  minimal: {
    pkg: Hi2,
    names: {
      Edit: 'HiPencil',
      Check: 'HiCheck',
      Trash2: 'HiTrash',
      Plus: 'HiPlus',
      Save: 'HiCloudArrowDown',
      X: 'HiXMark',
      Search: 'HiMagnifyingGlass',
      Maximize2: 'HiArrowsPointingOut',
      Minimize2: 'HiArrowsPointingIn',
      FileText: 'HiDocumentText',
      Moon: 'HiMoon',
      Sun: 'HiSun',
      Settings: 'HiAdjustmentsHorizontal',
      Folder: 'HiFolder',
      ChevronDown: 'HiChevronDown',
      ChevronRight: 'HiChevronRight',
      User: 'HiUser',
      Home: 'HiHome',
      Wrench: 'HiWrench',
      Star: 'HiStar',
      Zap: 'HiBolt',
      Code: 'HiCodeBracketSquare',
      MessageSquare: 'HiChatBubbleLeft',
      Sparkles: 'HiSparkles',
      Eye: 'HiEye',
      Network: 'HiCube',
    },
  },
  material: {
    pkg: Md,
    names: {
      Edit: 'MdEdit',
      Check: 'MdCheck',
      Trash2: 'MdDelete',
      Plus: 'MdAdd',
      Save: 'MdSave',
      X: 'MdClose',
      Search: 'MdSearch',
      Maximize2: 'MdFullscreen',
      Minimize2: 'MdFullscreenExit',
      FileText: 'MdDescription',
      Moon: 'MdDarkMode',
      Sun: 'MdLightMode',
      Settings: 'MdSettings',
      Folder: 'MdFolder',
      ChevronDown: 'MdKeyboardArrowDown',
      ChevronRight: 'MdKeyboardArrowRight',
      User: 'MdPerson',
      Home: 'MdHome',
      Lightbulb: 'MdLightbulb',
      Star: 'MdStar',
      Zap: 'MdBolt',
      Code: 'MdCode',
      Layers: 'MdLayers',
      MessageSquare: 'MdMessage',
      Sparkles: 'MdAutoAwesome',
      Eye: 'MdRemoveRedEye',
      Network: 'MdNetworkWifi3Bar',
      Wrench: 'MdBuild',
    },
  },
  rounded: {
    pkg: Ri,
    names: {
      Edit: 'RiPencilLine',
      Check: 'RiCheckLine',
      Trash2: 'RiDeleteBinLine',
      Plus: 'RiAddLine',
      Save: 'RiSaveLine',
      X: 'RiCloseLine',
      Search: 'RiSearchLine',
      Maximize2: 'RiFullscreenLine',
      Minimize2: 'RiFullscreenExitLine',
      FileText: 'RiFileTextLine',
      Moon: 'RiMoonLine',
      Sun: 'RiSunLine',
      Settings: 'RiSettingsLine',
      Folder: 'RiFolderLine',
      ChevronDown: 'RiArrowDownSLine',
      ChevronRight: 'RiArrowRightSLine',
      User: 'RiUserLine',
      Home: 'RiHomeLine',
      Star: 'RiStarLine',
      Zap: 'RiLightbulbFlashLine',
      Wrench: 'RiWrenchLine',
      Layers: 'RiStackLine',
      Code: 'RiCodeLine',
      MessageSquare: 'RiMessageLine',
      Sparkles: 'RiStarLine',
      Eye: 'RiEyeLine',
      Network: 'RiNetworkLine',
    },
  },
};

interface IconProps {
  name: string;
  className?: string;
  size?: number;
}

export function Icon({ name, className, size }: IconProps) {
  const { currentIconSet } = useTheme();
  const cfg = setConfig[currentIconSet] || setConfig.modern;
  const exportName = cfg.names[name];
  if (!exportName) return null;
  const Component = cfg.pkg[exportName] as ComponentType<any> | undefined;
  if (!Component) return null;

  return (
    <span
      className={cn(
        'inline-flex items-center justify-center shrink-0',
        className,
      )}
      style={size ? { width: size, height: size } : undefined}
    >
      <Component className="w-full h-full" />
    </span>
  );
}
