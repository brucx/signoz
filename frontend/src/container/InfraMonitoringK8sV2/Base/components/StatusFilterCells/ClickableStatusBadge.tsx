import { Badge, BadgeColor } from '@signozhq/ui/badge';

import { StatusFilterKind, useSetStatusFilter } from './useSetStatusFilter';

import styles from './ClickableStatusBadge.module.scss';

interface ClickableStatusBadgeProps {
	color: BadgeColor;
	label: string;
	/** The status this badge stands for, applied as the filter when clicked. */
	status: string;
	kind: StatusFilterKind;
}

function ClickableStatusBadge({
	color,
	label,
	status,
	kind,
}: ClickableStatusBadgeProps): JSX.Element {
	const setStatusFilter = useSetStatusFilter(kind);

	return (
		<button
			type="button"
			className={styles.statusBadgeButton}
			data-testid={`status-badge-${status}`}
			title={`Show only ${label}`}
			onClick={(e): void => {
				e.preventDefault();
				e.stopPropagation();
				setStatusFilter([status]);
			}}
		>
			<Badge color={color} variant="outline">
				{label}
			</Badge>
		</button>
	);
}

export default ClickableStatusBadge;
