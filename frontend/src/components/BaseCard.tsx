import { ReactNode, MouseEventHandler } from 'react';

interface BaseCardProps {
  color?: string;
  inactive?: boolean;
  selected?: boolean;
  small?: boolean;
  headerLeft?: ReactNode;
  headerRight?: ReactNode;
  icon?: ReactNode;
  title?: ReactNode;
  children?: ReactNode;
  className?: string;
  onClick?: MouseEventHandler<HTMLDivElement>;
}

export function BaseCard({
  color,
  inactive,
  selected,
  small,
  headerLeft,
  headerRight,
  icon,
  title,
  children,
  className = '',
  onClick,
}: BaseCardProps) {
  return (
    <div
      className={`base-card ${small ? 'base-card-sm' : ''} ${inactive ? 'inactive' : ''} ${selected ? 'selected' : ''} ${className}`}
      onClick={onClick}
    >
      <div className="base-card-header" style={{ backgroundColor: color }}>
        <div className="base-card-header-left">{headerLeft}</div>
        <div className="base-card-header-right">{headerRight}</div>
      </div>
      <div className="base-card-body">
        {icon && (
          <div className="base-card-icon" style={{ color }}>
            {icon}
          </div>
        )}
        {title && <div className="base-card-title">{title}</div>}
        {children}
      </div>
    </div>
  );
}
