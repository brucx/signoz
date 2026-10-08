import { findOptionLabel } from '../utils';

describe('findOptionLabel', () => {
	const sectioned = [
		{
			label: 'Lifecycle',
			options: [
				{ label: 'Running', value: 'running' },
				{ label: 'Pending', value: 'pending' },
			],
		},
		{
			label: 'Error status',
			options: [{ label: 'CrashLoopBackOff', value: 'crashloopbackoff' }],
		},
	];

	it('resolves a label nested inside a section', () => {
		expect(findOptionLabel(sectioned, 'crashloopbackoff')).toBe(
			'CrashLoopBackOff',
		);
	});

	it('resolves a label on a flat option list', () => {
		expect(findOptionLabel([{ label: 'Ready', value: 'ready' }], 'ready')).toBe(
			'Ready',
		);
	});

	it('returns undefined for a value no option declares', () => {
		expect(findOptionLabel(sectioned, 'nope')).toBeUndefined();
	});
});
