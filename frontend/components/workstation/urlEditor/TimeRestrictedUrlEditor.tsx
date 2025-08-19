import React, { useState, useEffect } from 'react';
import {
    Button,
    TextField,
    Select,
    Alert,
    Table,
    BodyShort,
    Label,
    VStack,
    HStack,
    Checkbox,
    Tooltip,
    Tag
} from '@navikt/ds-react';
import { TrashIcon, LinkIcon, PlusIcon } from '@navikt/aksel-icons';
import { formatDistanceToNow, addHours, isAfter } from 'date-fns';
import { nb } from 'date-fns/locale';
import { useWorkstationURLListForIdent, useCreateWorkstationURLListItemForIdent, useUpdateWorkstationURLListItemForIdent, useDeleteWorkstationURLListItemForIdent, useActivateWorkstationURLListForIdent } from '../queries';
import { WorkstationURLListItem } from '../../../lib/rest/generatedDto';

export interface TimeRestrictedUrl {
    id: string;
    url: string;
    description: string;
    duration: '12hours' | '1hour';
    createdAt: Date;
    expiresAt: Date;
    isExpired: boolean;
    hasChanges?: boolean;
    editingDescription?: string;
    editingDuration?: '12hours' | '1hour';
}

const TIME_DURATIONS = {
    '1hour': { label: '1 time', value: '1hour', hours: 1 },
    '12hours': { label: '12 timer', value: '12hours', hours: 12 }
} as const;

// Predefined description options in Norwegian
const PREDEFINED_DESCRIPTIONS = [
    'Koderepository',
    'Datakilde',
    'API',
    'Teknisk dokumentasjon',
    'Utviklingsverktøy',
    'Læringsressurser',
    'Feilsporing',
    'Skytjenester',
    'Annet'
];

// Use the same URL validation as the existing URL editor
const isValidUrl = (url: string) => {
    const urlPattern = /^((\*|\*?[a-zA-Z0-9-]+)\.)+[a-zA-Z0-9-]{2,}(\/(\*|[a-zA-Z0-9-._~:/?#[\]@!$&'()*+,;=]*))*$/;
    return !url ? true : urlPattern.test(url);
}

const TimeRestrictedUrlEditor: React.FC = () => {
    const { data: urlListData, isLoading: isLoadingData, refetch } = useWorkstationURLListForIdent();
    const createUrlMutation = useCreateWorkstationURLListItemForIdent();
    const updateUrlMutation = useUpdateWorkstationURLListItemForIdent();
    const deleteUrlMutation = useDeleteWorkstationURLListItemForIdent();
    const activateUrlsMutation = useActivateWorkstationURLListForIdent();

    const [timeRestrictedUrls, setTimeRestrictedUrls] = useState<TimeRestrictedUrl[]>([]);
    const [newUrl, setNewUrl] = useState('');
    const [newDescription, setNewDescription] = useState('');
    const [customDescription, setCustomDescription] = useState('');
    const [selectedDuration, setSelectedDuration] = useState<'12hours' | '1hour'>('1hour');
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);
    const [selectedUrls, setSelectedUrls] = useState<Set<string>>(new Set());
    const [showNewUrlForm, setShowNewUrlForm] = useState(false);

    // Transform backend data to frontend format
    const transformBackendData = (items: (WorkstationURLListItem | undefined)[]): TimeRestrictedUrl[] => {
        return items
            .filter((item): item is WorkstationURLListItem => item !== undefined)
            .map(item => {
                // Handle different duration formats from backend
                let duration: '12hours' | '1hour' = '1hour';
                if (item.duration) {
                    // Backend may return intervals like '1 hour', '12 hours', '01:00:00', '12:00:00', etc.
                    const durationStr = item.duration.toLowerCase().replace(/\s+/g, '');
                    if (durationStr.includes('12') || durationStr === '12:00:00') {
                        duration = '12hours';
                    } else if (durationStr.includes('1') || durationStr === '01:00:00') {
                        duration = '1hour';
                    }
                }

                return {
                    id: item.id || Date.now().toString(),
                    url: item.url || '',
                    description: item.description || '',
                    duration: duration,
                    createdAt: new Date(item.createdAt || Date.now()),
                    expiresAt: new Date(item.expiresAt || Date.now()),
                    isExpired: new Date() > new Date(item.expiresAt || Date.now())
                };
            });
    };

    // Load data from API
    useEffect(() => {
        if (urlListData?.items) {
            setTimeRestrictedUrls(transformBackendData(urlListData.items));
        }
    }, [urlListData]);

    // Update expired status periodically
    useEffect(() => {
        const interval = setInterval(() => {
            setTimeRestrictedUrls(prevUrls =>
                prevUrls.map(url => ({
                    ...url,
                    isExpired: isAfter(new Date(), url.expiresAt)
                }))
            );
        }, 60000); // Check every minute

        return () => clearInterval(interval);
    }, []);

    const handleAddUrl = async () => {
        if (!newUrl.trim()) {
            setError('URL kan ikke være tom');
            return;
        }

        if (!isValidUrl(newUrl)) {
            setError('Ugyldig URL format');
            return;
        }

        // Check if this exact URL is already active
        const existingActiveUrl = timeRestrictedUrls.find(url => url.url === newUrl.trim() && !url.isExpired);
        if (existingActiveUrl) {
            setError('Denne URL-en er allerede aktiv. Vent til den utløper eller fjern den først.');
            return;
        }

        setError(null);

        try {
            const now = new Date();
            const duration = TIME_DURATIONS[selectedDuration];
            const expiresAt = selectedDuration === '1hour'
                ? addHours(now, duration.hours)
                : addHours(now, duration.hours);

            // Use custom description if provided, otherwise use selected predefined description
            const finalDescription = customDescription.trim() || newDescription || 'Ingen beskrivelse';

            const newItem: Partial<WorkstationURLListItem> = {
                url: newUrl.trim(),
                description: finalDescription,
                duration: selectedDuration,
                createdAt: now.toISOString(),
                expiresAt: expiresAt.toISOString()
            };

            await createUrlMutation.mutateAsync(newItem as WorkstationURLListItem);

            setNewUrl('');
            setNewDescription('');
            setCustomDescription('');
            setShowNewUrlForm(false);
            setSuccess(`URL lagt til med ${duration.label} tilgang`);

            // Refresh the data
            refetch();

            // Clear success message after 3 seconds
            setTimeout(() => setSuccess(null), 3000);
        } catch (err) {
            setError('Kunne ikke legge til URL');
        }
    };

    const handleRemoveUrl = async (id: string) => {
        try {
            await deleteUrlMutation.mutateAsync(id);
            setSuccess('URL fjernet');

            // Refresh the data
            refetch();

            setTimeout(() => setSuccess(null), 3000);
        } catch (err) {
            setError('Kunne ikke fjerne URL');
            setTimeout(() => setError(null), 3000);
        }
    };

    const handleReopenExpiredUrl = async (url: TimeRestrictedUrl) => {
        try {
            const now = new Date();
            const duration = TIME_DURATIONS[url.duration];
            const expiresAt = addHours(now, duration.hours);

            const newItem: Partial<WorkstationURLListItem> = {
                url: url.url,
                description: url.description,
                duration: url.duration,
                createdAt: now.toISOString(),
                expiresAt: expiresAt.toISOString()
            };

            await createUrlMutation.mutateAsync(newItem as WorkstationURLListItem);

            setSuccess(`URL gjenåpnet med ${duration.label} tilgang`);

            // Refresh the data
            refetch();

            setTimeout(() => setSuccess(null), 3000);
        } catch (err) {
            setError('Kunne ikke gjenåpne URL');
        }
    };

    const handleUrlSelection = (urlId: string, checked: boolean) => {
        const newSelectedUrls = new Set(selectedUrls);
        if (checked) {
            newSelectedUrls.add(urlId);
        } else {
            newSelectedUrls.delete(urlId);
        }
        setSelectedUrls(newSelectedUrls);
    };

    const handleOpenSelectedUrls = async () => {
        const selectedUrlIds = Array.from(selectedUrls).filter(urlId => {
            const url = timeRestrictedUrls.find(u => u.id === urlId);
            return url && url.isExpired && url.description && url.description.trim() !== '';
        });

        if (selectedUrlIds.length === 0) {
            setError('Ingen utløpte URL-er med beskrivelse er valgt');
            setTimeout(() => setError(null), 5000);
            return;
        }

        try {
            await activateUrlsMutation.mutateAsync(selectedUrlIds);

            setSuccess(`Aktiverte ${selectedUrlIds.length} URL-er`);
            setTimeout(() => setSuccess(null), 5000);

            // Clear selection after activating
            setSelectedUrls(new Set());

            // Refresh the data
            refetch();
        } catch (err) {
            setError('Kunne ikke aktivere URL-er');
            setTimeout(() => setError(null), 5000);
        }
    };

    const handleOpenSingleUrl = (url: TimeRestrictedUrl) => {
        const fullUrl = url.url.startsWith('http') ? url.url : `https://${url.url}`;
        window.open(fullUrl, '_blank', 'noopener,noreferrer');
    };

    const handleStartEditing = (id: string) => {
        setTimeRestrictedUrls(prev =>
            prev.map(url =>
                url.id === id
                    ? {
                        ...url,
                        hasChanges: true,
                        editingDescription: url.description,
                        editingDuration: url.duration
                    }
                    : url
            )
        );
    };

    const handleCancelEditing = (id: string) => {
        setTimeRestrictedUrls(prev =>
            prev.map(url =>
                url.id === id
                    ? {
                        ...url,
                        hasChanges: false,
                        editingDescription: undefined,
                        editingDuration: undefined
                    }
                    : url
            )
        );
    };

    const handleDescriptionChange = (id: string, value: string) => {
        setTimeRestrictedUrls(prev =>
            prev.map(url => {
                if (url.id === id) {
                    const hasChanges = value !== url.description || (url.editingDuration && url.editingDuration !== url.duration);
                    return {
                        ...url,
                        editingDescription: value,
                        hasChanges: hasChanges
                    };
                }
                return url;
            })
        );
    };

    const handleDurationChange = (id: string, value: '12hours' | '1hour') => {
        setTimeRestrictedUrls(prev =>
            prev.map(url => {
                if (url.id === id) {
                    const hasChanges = value !== url.duration || (url.editingDescription && url.editingDescription !== url.description);
                    return {
                        ...url,
                        editingDuration: value,
                        hasChanges: !!hasChanges
                    };
                }
                return url;
            })
        );
    };

    const handleUpdateUrl = async (id: string) => {
        const url = timeRestrictedUrls.find(u => u.id === id);
        if (!url) return;

        const newDescription = url.editingDescription ?? url.description;
        const newDuration = url.editingDuration ?? url.duration;

        if (!newDescription.trim()) {
            setError('Beskrivelse kan ikke være tom');
            setTimeout(() => setError(null), 3000);
            return;
        }

        try {
            // Use the real PUT endpoint to update the URL item
            const now = new Date();
            const durationInfo = TIME_DURATIONS[newDuration];

            // Calculate new expiration time if duration changed
            let newExpiresAt = url.expiresAt;
            if (newDuration !== url.duration && !url.isExpired) {
                newExpiresAt = newDuration === '1hour'
                    ? addHours(now, durationInfo.hours)
                    : addHours(now, durationInfo.hours);
            }

            const updatedItem: WorkstationURLListItem = {
                id: url.id,
                url: url.url,
                description: newDescription,
                duration: newDuration,
                createdAt: new Date(url.createdAt).toISOString(),
                expiresAt: newExpiresAt.toISOString()
            };

            await updateUrlMutation.mutateAsync(updatedItem);

            setSuccess('URL oppdatert');

            // Refresh the data
            refetch();

            setTimeout(() => setSuccess(null), 3000);
        } catch (err) {
            setError('Kunne ikke oppdatere URL');
            setTimeout(() => setError(null), 3000);
        }
    };

    const handleEditingDescriptionChange = (id: string, value: string) => {
        setTimeRestrictedUrls(prev =>
            prev.map(url =>
                url.id === id
                    ? { ...url, editingDescription: value }
                    : url
            )
        );
    };

    const handleEditingDurationChange = (id: string, value: '12hours' | '1hour') => {
        setTimeRestrictedUrls(prev =>
            prev.map(url =>
                url.id === id
                    ? { ...url, editingDuration: value }
                    : url
            )
        );
    };

    const getTimeRemaining = (expiresAt: Date): string => {
        const now = new Date();
        if (isAfter(now, expiresAt)) {
            return 'Utløpt';
        }
        return `Utløper ${formatDistanceToNow(expiresAt, { addSuffix: true, locale: nb })}`;
    };

    const getStatusVariant = (url: TimeRestrictedUrl): 'success' | 'warning' | 'error' => {
        if (url.isExpired) return 'error';

        const timeLeft = url.expiresAt.getTime() - new Date().getTime();
        const oneHour = 60 * 60 * 1000;

        if (timeLeft < oneHour) return 'warning';
        return 'success';
    };

    const handleExtendUrl = async (url: TimeRestrictedUrl) => {
        try {
            // Use the activate endpoint to extend the URL
            await activateUrlsMutation.mutateAsync([url.id]);

            const duration = TIME_DURATIONS[url.duration];
            setSuccess(`URL utvidet med ${duration.label}`);

            // Refresh the data
            refetch();

            setTimeout(() => setSuccess(null), 3000);
        } catch (err) {
            setError('Kunne ikke utvide URL');
            setTimeout(() => setError(null), 3000);
        }
    };

    const isExpiringInLessThanOneHour = (url: TimeRestrictedUrl): boolean => {
        if (url.isExpired) return false;
        const timeLeft = url.expiresAt.getTime() - new Date().getTime();
        const oneHour = 60 * 60 * 1000;
        return timeLeft < oneHour;
    };

    const expiredUrls = timeRestrictedUrls.filter(url => url.isExpired);

    return (
        <div className="space-y-6">
            <div>
                <div className="flex flex-col items-center mb-6">
                    <div className="text-center mb-4">
                        <BodyShort className="text-gray-600">
                            Administrer URL-er som får midlertidig tilgang. URL-er må være uten https:// prefikset.
                        </BodyShort>
                    </div>
                    <Button
                        type="button"
                        variant="primary"
                        size="medium"
                        onClick={() => setShowNewUrlForm(!showNewUrlForm)}
                        icon={<PlusIcon aria-hidden />}
                        className={`px-8 py-3 text-lg ${showNewUrlForm ? 'hidden' : ''}`}
                    >
                        Ny URL
                    </Button>
                </div>

                {showNewUrlForm && (
                    <div className="flex justify-center mb-6">
                        <div className="bg-white p-6 rounded-lg shadow-sm border max-w-lg w-full">
                            <VStack gap="4">
                                <TextField
                                    label="URL"
                                    placeholder="example.com"
                                    value={newUrl}
                                    onChange={(e) => setNewUrl(e.target.value)}
                                    error={newUrl && !isValidUrl(newUrl) ? "Ugyldig URL format" : undefined}
                                    disabled={createUrlMutation.isPending || isLoadingData}
                                />

                                <Select
                                    label="Beskrivelse"
                                    value={newDescription}
                                    onChange={(e) => {
                                        setNewDescription(e.target.value);
                                        // Clear custom description when selecting predefined
                                        if (e.target.value && e.target.value !== 'custom') {
                                            setCustomDescription('');
                                        }
                                    }}
                                    disabled={createUrlMutation.isPending || isLoadingData}
                                >
                                    <option value="">Velg beskrivelse</option>
                                    {PREDEFINED_DESCRIPTIONS.map(desc => (
                                        <option key={desc} value={desc}>{desc}</option>
                                    ))}
                                    <option value="custom">Egendefinert...</option>
                                </Select>

                                {(newDescription === 'custom' || (!PREDEFINED_DESCRIPTIONS.includes(newDescription) && newDescription && newDescription !== 'custom')) && (
                                    <TextField
                                        label="Egendefinert beskrivelse"
                                        placeholder="Skriv din egen beskrivelse"
                                        value={customDescription}
                                        onChange={(e) => setCustomDescription(e.target.value)}
                                        disabled={createUrlMutation.isPending || isLoadingData}
                                    />
                                )}

                                <Select
                                    label="Varighet"
                                    value={selectedDuration}
                                    onChange={(e) => setSelectedDuration(e.target.value as '12hours' | '1hour')}
                                    disabled={createUrlMutation.isPending || isLoadingData}
                                >
                                    {Object.values(TIME_DURATIONS).map(duration => (
                                        <option key={duration.value} value={duration.value}>
                                            {duration.label}
                                        </option>
                                    ))}
                                </Select>

                                <HStack gap="2">
                                    <Button
                                        onClick={handleAddUrl}
                                        loading={createUrlMutation.isPending}
                                        disabled={!newUrl.trim() || !isValidUrl(newUrl)}
                                    >
                                        Legg til URL
                                    </Button>
                                    <Button
                                        variant="tertiary"
                                        onClick={() => setShowNewUrlForm(false)}
                                        disabled={createUrlMutation.isPending}
                                    >
                                        Avbryt
                                    </Button>
                                </HStack>
                            </VStack>
                        </div>
                    </div>
                )}
            </div>

            {error && (
                <Alert variant="error" className="max-w-lg mx-auto">
                    {error}
                </Alert>
            )}

            {success && (
                <Alert variant="success" className="max-w-lg mx-auto">
                    {success}
                </Alert>
            )}

            {/* URLs table with checkboxes */}
            {timeRestrictedUrls.length > 0 && (
                <div className="bg-white rounded-lg shadow-sm border">
                    <div className="flex items-center justify-between p-4 border-b bg-gray-50 rounded-t-lg">
                        <Label>Aktive og tidligere URL-tilganger</Label>
                        {expiredUrls.length > 0 && (
                            <Button
                                variant="secondary"
                                size="small"
                                onClick={handleOpenSelectedUrls}
                                disabled={selectedUrls.size === 0 || activateUrlsMutation.isPending}
                                loading={activateUrlsMutation.isPending}
                                icon={<LinkIcon aria-hidden />}
                            >
                                Aktiver valgte ({selectedUrls.size})
                            </Button>
                        )}
                    </div>

                    <div className="overflow-x-auto">
                        <Table className="w-full">
                            <Table.Header>
                                <Table.Row>
                                    <Table.HeaderCell scope="col" className="w-16">Velg</Table.HeaderCell>
                                    <Table.HeaderCell scope="col" className="w-64">URL og detaljer</Table.HeaderCell>
                                    <Table.HeaderCell scope="col" className="w-40">Status</Table.HeaderCell>
                                    <Table.HeaderCell scope="col" className="w-56">Handlinger</Table.HeaderCell>
                                </Table.Row>
                            </Table.Header>
                            <Table.Body>
                                {timeRestrictedUrls
                                    .sort((a, b) => {
                                        // Sort by expiry time descending (longest time before expiry first)
                                        return b.expiresAt.getTime() - a.expiresAt.getTime();
                                    })
                                    .map((url) => (
                                    <Table.Row key={url.id} className={url.isExpired ? 'bg-red-50/50' : isExpiringInLessThanOneHour(url) ? 'bg-yellow-50/50' : 'bg-green-50/20'}>
                                        <Table.DataCell className="w-16 align-top pt-4">
                                            {(!url.description || url.description.trim() === '') ? (
                                                <Tooltip content="Legg til en beskrivelse for å kunne åpne mot URL-en">
                                                    <div>
                                                        <Checkbox
                                                            checked={selectedUrls.has(url.id)}
                                                            onChange={(e) => handleUrlSelection(url.id, e.target.checked)}
                                                            disabled={!url.isExpired || !url.description || url.description.trim() === ''}
                                                            aria-label={`Velg ${url.url}`}
                                                            value={url.url}
                                                        >
                                                            {""}
                                                        </Checkbox>
                                                    </div>
                                                </Tooltip>
                                            ) : (
                                                <Checkbox
                                                    checked={selectedUrls.has(url.id)}
                                                    onChange={(e) => handleUrlSelection(url.id, e.target.checked)}
                                                    disabled={!url.isExpired || !url.description || url.description.trim() === ''}
                                                    aria-label={`Velg ${url.url}`}
                                                    value={url.url}
                                                >
                                                    {""}
                                                </Checkbox>
                                            )}
                                        </Table.DataCell>
                                        <Table.DataCell className="w-64">
                                            <div className="space-y-3">
                                                <div className="flex items-center gap-2">
                                                    <div className={`font-mono text-sm font-medium flex-1 ${url.isExpired ? 'text-gray-500' : 'text-gray-900'}`}>
                                                        {url.url}
                                                    </div>
                                                    {url.isExpired ? (
                                                        <Tag variant="error" size="small">Utløpt</Tag>
                                                    ) : isExpiringInLessThanOneHour(url) ? (
                                                        <Tag variant="warning" size="small">Utløper snart</Tag>
                                                    ) : (
                                                        <Tag variant="success" size="small">Aktiv</Tag>
                                                    )}
                                                </div>
                                                <div className="space-y-1">
                                                    <div className="text-xs text-gray-600">Beskrivelse:</div>
                                                    <Select
                                                        label=""
                                                        value={url.editingDescription ?? url.description}
                                                        onChange={(e) => handleDescriptionChange(url.id, e.target.value)}
                                                        disabled={createUrlMutation.isPending || isLoadingData}
                                                        className="w-full"
                                                        size="small"
                                                    >
                                                        <option value="">Velg beskrivelse</option>
                                                        {PREDEFINED_DESCRIPTIONS.map(desc => (
                                                            <option key={desc} value={desc}>{desc}</option>
                                                        ))}
                                                        {url.description && !PREDEFINED_DESCRIPTIONS.includes(url.description) && (
                                                            <option value={url.description}>{url.description}</option>
                                                        )}
                                                    </Select>
                                                </div>
                                                <div className="space-y-1">
                                                    <div className="text-xs text-gray-600">Varighet:</div>
                                                    <Select
                                                        label=""
                                                        value={url.editingDuration ?? url.duration}
                                                        onChange={(e) => handleDurationChange(url.id, e.target.value as '12hours' | '1hour')}
                                                        disabled={createUrlMutation.isPending || isLoadingData}
                                                        className="w-full"
                                                        size="small"
                                                    >
                                                        {Object.values(TIME_DURATIONS).map(duration => (
                                                            <option key={duration.value} value={duration.value}>
                                                                {duration.label}
                                                            </option>
                                                        ))}
                                                    </Select>
                                                </div>
                                            </div>
                                        </Table.DataCell>
                                        <Table.DataCell className="w-40 align-top pt-4">
                                            <div className="space-y-2">
                                                <div className="flex items-center space-x-2">
                                                    {url.isExpired ? (
                                                        <>
                                                            <div className="w-2 h-2 bg-red-500 rounded-full flex-shrink-0"></div>
                                                            <span className="text-red-700 text-sm font-medium">
                                                                Utløpt {formatDistanceToNow(url.expiresAt, { addSuffix: true, locale: nb })}
                                                            </span>
                                                        </>
                                                    ) : isExpiringInLessThanOneHour(url) ? (
                                                        <>
                                                            <div className="w-2 h-2 bg-yellow-500 rounded-full flex-shrink-0"></div>
                                                            <span className="text-yellow-700 text-sm font-medium">
                                                                Utløper {formatDistanceToNow(url.expiresAt, { addSuffix: true, locale: nb })}
                                                            </span>
                                                        </>
                                                    ) : (
                                                        <>
                                                            <div className="w-2 h-2 bg-green-500 rounded-full flex-shrink-0"></div>
                                                            <span className="text-green-700 text-sm font-medium">
                                                                Aktiv, utløper {formatDistanceToNow(url.expiresAt, { addSuffix: true, locale: nb })}
                                                            </span>
                                                        </>
                                                    )}
                                                </div>
                                            </div>
                                        </Table.DataCell>
                                        <Table.DataCell className="w-56 align-top pt-4 h-full min-h-32">
                                            <div className="flex flex-col gap-2 h-full justify-start">
                                                {/* For expired URLs: Oppdater, Fjern */}
                                                {url.isExpired ? (
                                                    <>
                                                        {/* Oppdater button (top for expired URLs) */}
                                                        <Button
                                                            variant="secondary"
                                                            size="small"
                                                            onClick={() => handleUpdateUrl(url.id)}
                                                            disabled={!url.hasChanges}
                                                            className="w-full"
                                                        >
                                                            Oppdater
                                                        </Button>

                                                        {/* Fjern button (bottom for expired URLs) */}
                                                        <Button
                                                            variant="danger"
                                                            size="small"
                                                            onClick={() => handleRemoveUrl(url.id)}
                                                            icon={<TrashIcon aria-hidden />}
                                                            className="w-full"
                                                        >
                                                            Fjern
                                                        </Button>
                                                    </>
                                                ) : (
                                                    <>
                                                        {/* For active URLs that expire in less than 1 hour: show Extend button */}
                                                        {isExpiringInLessThanOneHour(url) && (
                                                            <Button
                                                                variant="primary"
                                                                size="small"
                                                                onClick={() => handleExtendUrl(url)}
                                                                disabled={activateUrlsMutation.isPending}
                                                                className="w-full"
                                                            >
                                                                Utvid ({TIME_DURATIONS[url.duration].label})
                                                            </Button>
                                                        )}

                                                        {/* Oppdater button for active URLs */}
                                                        <Button
                                                            variant="secondary"
                                                            size="small"
                                                            onClick={() => handleUpdateUrl(url.id)}
                                                            disabled={!url.hasChanges}
                                                            className="w-full"
                                                        >
                                                            Oppdater
                                                        </Button>

                                                        {/* Fjern button for active URLs */}
                                                        <Button
                                                            variant="danger"
                                                            size="small"
                                                            onClick={() => handleRemoveUrl(url.id)}
                                                            icon={<TrashIcon aria-hidden />}
                                                            className="w-full"
                                                        >
                                                            Fjern
                                                        </Button>
                                                    </>
                                                )}
                                            </div>
                                        </Table.DataCell>
                                    </Table.Row>
                                ))}
                            </Table.Body>
                        </Table>
                    </div>
                </div>
            )}

            {timeRestrictedUrls.length === 0 && !isLoadingData && (
                <div className="text-center py-8">
                    <BodyShort className="text-gray-500">
                        Ingen URL-tilganger registrert. Klikk &quot;Ny URL&quot;for å legge til din første tidsbegrensede åpning.
                    </BodyShort>
                </div>
            )}
        </div>
    );
};

export default TimeRestrictedUrlEditor;
