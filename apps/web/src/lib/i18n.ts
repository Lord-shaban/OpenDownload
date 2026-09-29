export type Locale = "en" | "ar";
export const LOCALE_COOKIE = "od_locale";
export function parseLocale(value?: string): Locale {
  return value === "ar" ? "ar" : "en";
}
export function direction(locale: Locale) {
  return locale === "ar" ? "rtl" : "ltr";
}
// Each product message has both translations; keys and placeholders are checked.
export const messages = {
  skip: ["Skip to workspace", "انتقل إلى مساحة العمل"],
  home: ["OpenDownload home", "الصفحة الرئيسية لـOpenDownload"],
  yourWorkspace: ["YOUR WORKSPACE", "مساحة عملك"],
  workspace: ["Workspace", "مساحة العمل"],
  newDownload: ["New download", "تنزيل جديد"],
  downloads: ["Downloads", "التنزيلات"],
  mainNav: ["Main navigation", "التنقل الرئيسي"],
  mobileNav: ["Mobile navigation", "التنقل على الهاتف"],
  simple: ["Made to stay simple.", "بساطة تدوم."],
  yourInstance: ["Open source. Your instance.", "مفتوح المصدر. على خادمك."],
  moreControl: [
    "A little more control over your media.",
    "تحكّم أكبر في الوسائط التي تحفظها.",
  ],
  helpGuidelines: ["Help & guidelines", "المساعدة والإرشادات"],
  github: ["View on GitHub", "عرض المشروع على GitHub"],
  preRelease: ["SELF-HOSTED / PRE-RELEASE", "استضافة ذاتية / نسخة تجريبية"],
  testInstance: ["Test instance", "بيئة اختبار"],
  instanceReady: ["Instance ready", "الخادم جاهز"],
  connecting: ["Connecting", "جارٍ الاتصال"],
  setupNeeded: ["Setup needed", "يلزم إكمال الإعداد"],
  lightTheme: ["Switch to light theme", "التبديل إلى المظهر الفاتح"],
  darkTheme: ["Switch to dark theme", "التبديل إلى المظهر الداكن"],
  language: ["Language", "اللغة"],
  help: ["Help", "المساعدة"],
  interruptedWorkspace: [
    "Something interrupted the workspace.",
    "تعذر عرض مساحة العمل.",
  ],
  reconnectWorkspace: [
    "Your queued downloads are stored on the server. Reload to reconnect.",
    "تنزيلاتك المنتظرة محفوظة على الخادم. حاول مجددًا لإعادة الاتصال.",
  ],
  tryAgain: ["Try again", "حاول مجددًا"],
  pageMissing: ["This page isn’t here.", "هذه الصفحة غير موجودة."],
  returnHome: ["Return to OpenDownload", "العودة إلى OpenDownload"],
  fixtureMode: ["Fixture mode.", "وضع الاختبار."],
  fixtureExplanation: [
    "This instance tests the complete workflow. It produces a labeled test file, not extracted media.",
    "تختبر هذه البيئة خطوات التنزيل كاملة، وتنتج ملف اختبار موضّحًا بدلًا من استخراج وسائط فعلية.",
  ],
  missingTools: [
    "Media tools are missing from this instance. Install yt-dlp and FFmpeg, or start the Docker deployment described in the README.",
    "أدوات الوسائط غير متوفرة على هذا الخادم. ثبّت yt-dlp وFFmpeg، أو شغّل نسخة Docker وفق تعليمات README.",
  ],
  connectInstance: ["Connecting to your instance…", "جارٍ الاتصال بخادمك…"],
  interrupted: [
    "Connection interrupted. Reconnecting; your jobs stay on the server.",
    "انقطع الاتصال. نحاول الاتصال مجددًا؛ مهامك محفوظة على الخادم.",
  ],
  freedom: ["LESS FRICTION. MORE FREEDOM.", "خطوات أقل. حرية أكبر."],
  library: ["YOUR TEMPORARY LIBRARY", "مكتبتك المؤقتة"],
  saveGood: ["Save something good.", "احفظ ما يستحق."],
  readyWhen: ["Ready when you are.", "جاهزة وقتما تريد."],
  intro: [
    "A link is all you need. Choose what to keep, and we’ll handle the rest.",
    "كل ما تحتاجه رابط. اختر ما تريد حفظه، وسنتولى الباقي.",
  ],
  downloadsIntro: [
    "Follow your downloads and save finished files before they expire.",
    "تابع تنزيلاتك واحفظ الملفات المكتملة قبل انتهاء صلاحيتها.",
  ],
  mediaLink: ["Media link", "رابط الوسائط"],
  placeholder: [
    "Paste a public video, audio, or image link",
    "ألصق رابط فيديو أو صوت أو صورة عامة",
  ],
  clearLink: ["Clear link", "مسح الرابط"],
  pasteLink: ["Paste link from clipboard", "لصق الرابط من الحافظة"],
  pasteClipboard: ["Paste from clipboard", "لصق من الحافظة"],
  analyzing: ["Analyzing…", "جارٍ التحليل…"],
  analyzeLink: ["Analyze link", "تحليل الرابط"],
  detected: ["{source} detected", "المصدر: {source}"],
  explicitAnalysis: [
    "Only analyzed when you ask. No login required.",
    "يبدأ التحليل عندما تطلبه. لا يلزم تسجيل الدخول.",
  ],
  cancelAnalysis: ["Cancel analysis", "إلغاء التحليل"],
  focusLink: ["Focus link", "انتقل إلى الرابط"],
  popularSources: ["POPULAR SOURCES", "مصادر شائعة"],
  otherSources: ["& other public sources", "ومصادر عامة أخرى"],
  looking: ["Looking at the source.", "جارٍ فحص المصدر."],
  checkingFormats: [
    "Checking public access and available formats…",
    "نتحقق من الوصول العام والصيغ المتاحة…",
  ],
  takesMoment: [
    "This can take a moment. You can cancel at any time.",
    "قد يستغرق ذلك قليلًا. يمكنك الإلغاء في أي وقت.",
  ],
  results: ["Analysis results", "نتائج التحليل"],
  howWorks: ["How it works", "كيف يعمل"],
  oneLink: ["One link.", "رابط واحد."],
  pastePublic: ["Paste a public media URL.", "ألصق رابط وسائط عامة."],
  yourChoice: ["Your choice.", "الخيار لك."],
  pickFormats: ["Pick from the available formats.", "اختر من الصيغ المتاحة."],
  yoursToSave: ["It’s yours to save.", "جاهز للحفظ."],
  processDownload: ["We process. You download.", "نعالج الملف، ثم تحفظه."],
  yourDownloads: ["Your downloads", "تنزيلاتك"],
  allDownloads: ["All downloads", "كل التنزيلات"],
  activeCount: ["{count} active", "قيد التنفيذ: {count}"],
  testLink: ["Try a test link", "جرّب رابط اختبار"],
  retention: ["Files kept for {hours}h", "تُحفظ الملفات لمدة {hours} س"],
  temporaryStorage: ["Temporary storage", "تخزين مؤقت"],
  noTrackers: [
    "No trackers. No ads. Just your media.",
    "بلا تتبّع أو إعلانات. وسائطك فقط.",
  ],
  perJob: ["{size} per job", "{size} لكل مهمة"],
  publicOnly: ["Public content only", "محتوى عام فقط"],
  deleteQuestion: ["Delete this download?", "حذف هذا التنزيل؟"],
  deleteExplanation: [
    "The job and its temporary files will be removed from this instance. Files already saved to your device are unaffected.",
    "ستُحذف المهمة وملفاتها المؤقتة من الخادم. الملفات التي حفظتها على جهازك تبقى كما هي.",
  ],
  keepDownload: ["Keep download", "الاحتفاظ بالتنزيل"],
  deleting: ["Deleting…", "جارٍ الحذف…"],
  deleteDownload: ["Delete download", "حذف التنزيل"],
  clipboardError: [
    "Clipboard access isn’t available. Paste the link into the field with Ctrl+V or ⌘V.",
    "تعذر الوصول إلى الحافظة. ألصق الرابط في الحقل باستخدام Ctrl+V أو ⌘V.",
  ],
  invalidUrl: [
    "Enter a complete public link starting with https:// or http://.",
    "أدخل رابطًا عامًا كاملًا يبدأ بـhttps:// أو http://.",
  ],
  analysisComplete: [
    "Analysis complete. Choose an available format.",
    "اكتمل التحليل. اختر صيغة متاحة.",
  ],
  queuedNotice: [
    "Download added to your queue.",
    "أُضيف التنزيل إلى قائمة الانتظار.",
  ],
  cancelNotice: ["Cancellation requested.", "أُرسل طلب الإلغاء."],
  retryNotice: [
    "A new retry was added to the queue.",
    "أُضيفت محاولة جديدة إلى قائمة الانتظار.",
  ],
  deleteNotice: [
    "Job and temporary files deleted.",
    "حُذفت المهمة وملفاتها المؤقتة.",
  ],
  analysisCanceled: ["Analysis canceled.", "أُلغي التحليل."],
  availableMedia: ["Available media", "الوسائط المتاحة"],
  testFixture: ["Test fixture", "عينة اختبار"],
  mediaType: ["Media type", "نوع الوسائط"],
  video: ["Video", "فيديو"],
  audio: ["Audio", "صوت"],
  images: ["Images", "صور"],
  subtitles: ["Subtitles", "ترجمة"],
  quality: ["AVAILABLE QUALITY", "الجودة المتاحة"],
  audioFormat: ["AUDIO FORMAT", "صيغة الصوت"],
  captionLanguage: ["CAPTION LANGUAGE", "لغة الترجمة"],
  imageFormat: ["IMAGE FORMAT", "صيغة الصورة"],
  fromSource: ["From the source", "من المصدر"],
  bestAvailable: ["BEST AVAILABLE", "أفضل جودة متاحة"],
  exceedsLimit: ["Exceeds instance limit", "يتجاوز حد الخادم"],
  testFileExplanation: [
    "A test file verifies this flow. No real media is extracted.",
    "ملف اختبار للتحقق من هذه الخطوات. لا تُستخرج وسائط فعلية.",
  ],
  temporaryExplanation: [
    "Saved temporarily on your instance. Original watermarks are preserved.",
    "يُحفظ مؤقتًا على خادمك. تبقى العلامات المائية الأصلية.",
  ],
  addingQueue: ["Adding to queue…", "جارٍ الإضافة…"],
  queueDownload: ["Queue download", "إضافة إلى قائمة الانتظار"],
  emptyTitle: ["A little space for what you save.", "مساحة لما تريد حفظه."],
  emptyExplanation: [
    "Your queued and completed downloads will appear here. Start with a public link above.",
    "ستظهر مهامك المنتظرة والمكتملة هنا. ابدأ برابط عام في الحقل أعلاه.",
  ],
  downloadJobs: ["Download jobs", "مهام التنزيل"],
  preparing: ["Preparing file", "جارٍ تجهيز الملف"],
  checkingSource: ["Checking source", "جارٍ فحص المصدر"],
  downloading: ["Downloading", "جارٍ التنزيل"],
  readySave: ["Ready to save", "جاهز للحفظ"],
  failed: ["Couldn’t finish", "تعذر إكمال التنزيل"],
  canceled: ["Canceled", "أُلغي"],
  inQueue: ["In queue", "في قائمة الانتظار"],
  expired: ["Expired", "انتهت الصلاحية"],
  hoursRemaining: ["{count}h remaining", "الوقت المتبقي: {count} س"],
  minutesRemaining: ["{count}m remaining", "الوقت المتبقي: {count} د"],
  cancelJob: ["Cancel {title}", "إلغاء {title}"],
  retryJob: ["Retry {title}", "إعادة محاولة {title}"],
  deleteJob: ["Delete {title} and its files", "حذف {title} وملفاته"],
  saveTestFile: ["Save test file", "حفظ ملف الاختبار"],
  saveNumberedFile: ["Save file {number}", "حفظ الملف {number}"],
  saveFile: ["Save file", "حفظ الملف"],
  helpTitle: ["A few things worth knowing.", "معلومات تساعدك على البدء."],
  helpIntro: [
    "OpenDownload saves media you own or have permission to download.",
    "يحفظ OpenDownload الوسائط التي تملكها أو لديك إذن بتنزيلها.",
  ],
  publicChoices: ["Public links, actual choices", "روابط عامة وخيارات فعلية"],
  publicChoicesExplanation: [
    "The extractor checks each link and returns the available formats. Platform support can change. Photo gallery coverage is currently limited to direct images and collections returned by the extractor.",
    "يفحص المستخرج كل رابط ويعرض الصيغ المتاحة. قد يتغير دعم المنصات. دعم معارض الصور حاليًا يقتصر على الصور المباشرة والمجموعات التي يعيدها المستخرج.",
  ],
  instanceFiles: ["Your instance, temporary files", "خادمك وملفات مؤقتة"],
  instanceFilesExplanation: [
    "Processing happens on your server. Files expire automatically; save them to your device before the expiry shown. Delete removes a job and its server files.",
    "تُعالج الملفات على خادمك وتنتهي صلاحيتها تلقائيًا. احفظها على جهازك قبل الوقت الموضّح. الحذف يزيل المهمة وملفاتها من الخادم.",
  ],
  accessBoundary: ["Access boundaries stay in place.", "نحترم حدود الوصول."],
  accessExplanation: [
    "No private or paid content, login cookies, DRM bypass, live streams, or watermark removal. Use one public media link at a time.",
    "لا ندعم المحتوى الخاص أو المدفوع، أو ملفات تسجيل الدخول، أو تجاوز DRM، أو البث المباشر، أو إزالة العلامات المائية. استخدم رابط وسائط عامة واحدًا في كل مرة.",
  ],
  readDocs: ["Read the project documentation", "اقرأ توثيق المشروع"],
  close: ["Close", "إغلاق"],
  sizeVaries: ["Size varies", "الحجم متغير"],
  originalVideo: ["Original video", "الفيديو الأصلي"],
  originalAudio: ["Original audio", "الصوت الأصلي"],
  originalImage: ["Original image", "الصورة الأصلية"],
  mp3Audio: ["MP3 audio", "صوت MP3"],
  thumbnail: ["Thumbnail", "الصورة المصغّرة"],
  originalImages: ["{count} original images", "صور أصلية: {count}"],
  boundedCollection: ["Bounded image collection", "مجموعة صور ضمن الحدود"],
  unchanged: ["Source file, unchanged", "الملف الأصلي دون تغيير"],
  sourceFile: ["Source file", "ملف المصدر"],
  originalSource: ["Original source file", "ملف المصدر الأصلي"],
  qualityUnknown: ["quality not reported", "الجودة غير محددة"],
  codecUnknown: ["codec not reported", "الترميز غير محدد"],
  audioIncluded: ["audio included", "مع الصوت"],
  originalNoConversion: [
    "Source audio, no conversion",
    "الصوت الأصلي دون تحويل",
  ],
  converted: ["Converted", "محوّل"],
  convertedSource: ["Converted from source", "محوّل من المصدر"],
  upToBitrate: ["up to 192 kbps", "حتى 192 kbps"],
  requiresAudio: [
    "Conversion · requires an audio track",
    "التحويل يتطلب مسارًا صوتيًا",
  ],
  coverImage: ["Original cover image", "صورة الغلاف الأصلية"],
  originalCaptions: ["Original captions", "الترجمة الأصلية"],
  automaticCaptions: ["Auto-generated", "مولّدة تلقائيًا"],
  unexpectedResponse: [
    "The server returned an unexpected response. Refresh to reconnect.",
    "أعاد الخادم استجابة غير متوقعة. حدّث الصفحة لإعادة الاتصال.",
  ],
  requestFailed: [
    "The server could not complete this request. Try again shortly.",
    "تعذر على الخادم إكمال الطلب. حاول مجددًا بعد قليل.",
  ],
  networkError: [
    "Could not connect to your instance. Check the connection and try again.",
    "تعذر الاتصال بخادمك. تحقق من الاتصال وحاول مجددًا.",
  ],
  originDenied: [
    "This request did not come from the configured OpenDownload origin.",
    "لم يأتِ الطلب من عنوان OpenDownload المُعدّ على الخادم.",
  ],
  invalidRequest: [
    "Check the link or choose an available format and try again.",
    "تحقق من الرابط أو اختر صيغة متاحة وحاول مجددًا.",
  ],
  rateLimited: [
    "Too many requests. Try again in a minute.",
    "طلبات كثيرة. حاول مجددًا بعد دقيقة.",
  ],
  unsafeUrl: [
    "Use a public HTTP or HTTPS URL without credentials or custom ports.",
    "استخدم رابط HTTP أو HTTPS عامًا دون بيانات دخول أو منافذ مخصصة.",
  ],
  analysisBusy: [
    "Analysis is busy. Try again in a moment.",
    "خدمة التحليل مشغولة. حاول مجددًا بعد قليل.",
  ],
  analysisTimeout: [
    "Analysis took too long or was canceled. Try a shorter public link.",
    "استغرق التحليل وقتًا طويلًا أو أُلغي. جرّب رابطًا عامًا أقصر.",
  ],
  sourceUnavailable: [
    "The source is unavailable or does not offer supported public media. Check the link or try again later. Private, authenticated and DRM-protected content is unsupported.",
    "المصدر غير متاح أو لا يوفر وسائط عامة مدعومة. تحقق من الرابط أو حاول لاحقًا. لا ندعم المحتوى الخاص أو المحمي أو الذي يتطلب تسجيل الدخول.",
  ],
  analysisExpired: [
    "This analysis expired. Analyze the link again.",
    "انتهت صلاحية التحليل. حلّل الرابط مجددًا.",
  ],
  fileTooLarge: [
    "This source exceeds the configured file limit. Choose a smaller format.",
    "يتجاوز المصدر حد حجم الملفات. اختر صيغة أصغر.",
  ],
  invalidFormat: [
    "This format is not offered by the analysis.",
    "هذه الصيغة ليست ضمن خيارات التحليل.",
  ],
  jobExpired: [
    "This job expired. Analyze the link again.",
    "انتهت صلاحية المهمة. حلّل الرابط مجددًا.",
  ],
  notFound: ["Job not found.", "المهمة غير موجودة."],
  queueFull: [
    "The queue is full. Wait for a download to finish and try again.",
    "قائمة الانتظار ممتلئة. انتظر اكتمال تنزيل ثم حاول مجددًا.",
  ],
  invalidState: [
    "The job changed or cancellation is finishing. Refresh and try again.",
    "تغيّرت حالة المهمة أو لم يكتمل الإلغاء بعد. حدّث الصفحة وحاول مجددًا.",
  ],
  fileUnavailable: [
    "The file is not ready or has expired.",
    "الملف غير جاهز أو انتهت صلاحيته.",
  ],
  jobFailure: [
    "This download could not finish. Analyze the public link again, choose a smaller format, or retry later.",
    "تعذر إكمال هذا التنزيل. حلّل الرابط العام مجددًا، أو اختر صيغة أصغر، أو حاول لاحقًا.",
  ],
} as const;
export type MessageKey = keyof typeof messages;
export function translate(
  locale: Locale,
  key: MessageKey,
  values: Record<string, string | number> = {},
) {
  return messages[key][locale === "ar" ? 1 : 0].replace(
    /\{(\w+)\}/g,
    (token, name: string) => String(values[name] ?? token),
  );
}
export function isolate(value: string) {
  return `\u2068${value}\u2069`;
}
