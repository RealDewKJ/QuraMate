export interface MarkdownInlineSegment {
    type: 'text' | 'image';
    text?: string;
    alt?: string;
    url?: string;
}

export interface MarkdownEmbeddedImage {
    id: string;
    alt: string;
    fileName: string;
    mimeType: string;
    dataUrl: string;
}

export interface MarkdownPreviewBlock {
    type: 'heading' | 'paragraph' | 'list' | 'code';
    level?: number;
    text?: string;
    lines?: string[];
    segments?: MarkdownInlineSegment[];
}

const MARKDOWN_IMAGE_PATTERN = /!\[([^\]]*)\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g;
const EMBEDDED_IMAGE_PREFIX = 'quramate-image:';

export const buildMarkdownImageMarkup = (altText: string, url: string): string => {
    const safeAltText = altText.trim();
    return `![${safeAltText}](${url})`;
};

export const buildEmbeddedImageUrl = (imageId: string): string => {
    return `${EMBEDDED_IMAGE_PREFIX}${imageId}`;
};

export const isEmbeddedImageUrl = (value: string): boolean => {
    return value.startsWith(EMBEDDED_IMAGE_PREFIX);
};

export const getEmbeddedImageId = (value: string): string => {
    return value.slice(EMBEDDED_IMAGE_PREFIX.length);
};

export const makeImageId = (): string => {
    return `image_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 10)}`;
};

export const parseInlineMarkdown = (value: string): MarkdownInlineSegment[] => {
    const segments: MarkdownInlineSegment[] = [];
    let lastIndex = 0;

    for (const match of value.matchAll(MARKDOWN_IMAGE_PATTERN)) {
        const matchIndex = match.index ?? 0;
        if (matchIndex > lastIndex) {
            segments.push({
                type: 'text',
                text: value.slice(lastIndex, matchIndex),
            });
        }

        segments.push({
            type: 'image',
            alt: match[1] || 'image',
            url: match[2],
        });
        lastIndex = matchIndex + match[0].length;
    }

    if (lastIndex < value.length) {
        segments.push({
            type: 'text',
            text: value.slice(lastIndex),
        });
    }

    return segments.length > 0 ? segments : [{ type: 'text', text: value }];
};

export const parseMarkdownPreviewBlocks = (content: string): MarkdownPreviewBlock[] => {
    const blocks: MarkdownPreviewBlock[] = [];
    const lines = content.replace(/\r\n/g, '\n').split('\n');
    let index = 0;

    while (index < lines.length) {
        const rawLine = lines[index];
        const line = rawLine.trimEnd();
        const trimmed = line.trim();

        if (!trimmed) {
            index += 1;
            continue;
        }

        if (trimmed.startsWith('```')) {
            const codeLines: string[] = [];
            index += 1;
            while (index < lines.length && !lines[index].trim().startsWith('```')) {
                codeLines.push(lines[index]);
                index += 1;
            }
            if (index < lines.length) {
                index += 1;
            }
            blocks.push({
                type: 'code',
                lines: codeLines.length > 0 ? codeLines : [''],
            });
            continue;
        }

        const headingMatch = trimmed.match(/^(#{1,4})\s+(.*)$/);
        if (headingMatch) {
            blocks.push({
                type: 'heading',
                level: headingMatch[1].length,
                text: headingMatch[2],
            });
            index += 1;
            continue;
        }

        if (/^[-*]\s+/.test(trimmed) || /^\d+\.\s+/.test(trimmed)) {
            const items: string[] = [];
            while (index < lines.length) {
                const currentTrimmed = lines[index].trim();
                if (!(/^[-*]\s+/.test(currentTrimmed) || /^\d+\.\s+/.test(currentTrimmed))) {
                    break;
                }
                items.push(currentTrimmed.replace(/^[-*]\s+/, '').replace(/^\d+\.\s+/, ''));
                index += 1;
            }
            blocks.push({
                type: 'list',
                lines: items,
            });
            continue;
        }

        const paragraphLines: string[] = [];
        while (index < lines.length && lines[index].trim()) {
            const paragraphTrimmed = lines[index].trim();
            if (
                paragraphTrimmed.startsWith('```') ||
                /^(#{1,4})\s+/.test(paragraphTrimmed) ||
                /^[-*]\s+/.test(paragraphTrimmed) ||
                /^\d+\.\s+/.test(paragraphTrimmed)
            ) {
                break;
            }
            paragraphLines.push(paragraphTrimmed);
            index += 1;
        }
        const paragraphText = paragraphLines.join(' ');
        blocks.push({
            type: 'paragraph',
            lines: paragraphLines,
            segments: parseInlineMarkdown(paragraphText),
        });
    }

    return blocks;
};

export const extractDataUrlImages = (content: string): { content: string; images: MarkdownEmbeddedImage[] } => {
    const images: MarkdownEmbeddedImage[] = [];
    let imageIndex = 0;
    const nextContent = content.replace(/!\[([^\]]*)\]\((data:image\/[^)\s]+)\)/g, (_match, altText: string, dataUrl: string) => {
        const imageId = makeImageId();
        imageIndex += 1;
        images.push({
            id: imageId,
            alt: altText || 'image',
            fileName: `image-${imageIndex}`,
            mimeType: dataUrl.slice(5, dataUrl.indexOf(';')) || 'image/png',
            dataUrl,
        });
        return buildMarkdownImageMarkup(altText || 'image', buildEmbeddedImageUrl(imageId));
    });

    return {
        content: nextContent,
        images,
    };
};

export const resolveInlineImageUrl = (
    url: string,
    embeddedImages: MarkdownEmbeddedImage[] | undefined,
): string => {
    if (!isEmbeddedImageUrl(url)) {
        return url;
    }

    const imageId = getEmbeddedImageId(url);
    return embeddedImages?.find((image) => image.id === imageId)?.dataUrl || '';
};

export const fileToDataUrl = (file: File): Promise<string> => {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => {
            const result = typeof reader.result === 'string' ? reader.result : '';
            if (!result) {
                reject(new Error('Unable to read image file.'));
                return;
            }
            resolve(result);
        };
        reader.onerror = () => {
            reject(new Error('Unable to read image file.'));
        };
        reader.readAsDataURL(file);
    });
};

export const isImageFile = (file: File): boolean => {
    return file.type.startsWith('image/');
};

export const getImageAltText = (fileName: string): string => {
    const withoutExtension = fileName.replace(/\.[^.]+$/, '');
    const normalized = withoutExtension.trim().replace(/[_-]+/g, ' ');
    return normalized.length > 0 ? normalized : 'image';
};
