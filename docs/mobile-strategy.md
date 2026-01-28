# Mobile Strategy: PWA vs React Native

_Note:_ This document compares Progressive Web App (PWA) and React Native approaches specifically for the Family Photo Management System, providing recommendations for V1 and future mobile enhancements.

## Executive Summary

**Recommendation for V1:** Build a Progressive Web App (PWA) using the existing Vite + React stack with mobile-optimized features and offline support.

**Rationale:**
- Single codebase reduces development time and maintenance
- PWA provides 90% of native features needed for photo management
- No app store approval process or distribution overhead
- Can upgrade to React Native in V2+ if native features become critical

**Future Consideration:** Evaluate React Native after V1 launch based on user feedback about performance, offline capabilities, and feature gaps.

---

## Detailed Comparison

### PWA (Progressive Web App)

#### What It Provides

**Installation & Access:**
- Add to home screen on iOS and Android
- Standalone app experience (fullscreen, no browser chrome)
- App icon on device launcher
- Splash screen on launch

**Device Features:**
- Camera access via `<input type="file" capture="camera">`
- Photo library access via file picker
- Geolocation API
- Push notifications (Android only, iOS 16.4+)
- Background sync (limited)
- Clipboard access
- Share API

**Offline Capabilities:**
- Service Worker caching for app shell
- IndexedDB for local data storage
- Queue uploads for when connection restored
- Cache thumbnails for offline browsing

**Performance:**
- Lazy load routes with code splitting
- Image lazy loading and virtualization
- Service Worker caching reduces network requests
- Good performance on modern mobile browsers

#### Limitations

**iOS Restrictions:**
- Push notifications limited (iOS 16.4+, Safari only)
- Background processing very limited
- 50MB storage quota for IndexedDB
- Camera quality lower than native
- No access to photo metadata (EXIF) from library

**General Limitations:**
- No true background upload sync
- Can't organize device photo library
- Less smooth animations than native
- No access to system photo picker with metadata
- Cannot integrate with system share sheet deeply

#### Development Requirements

**Technology:**
- Existing Vite + React stack
- `vite-plugin-pwa` for PWA features
- Workbox for service worker management
- IndexedDB (idb) for offline storage

**Effort Estimate:**
- 2-3 weeks to add PWA features to existing app
- Minimal ongoing maintenance
- Single codebase to maintain

**Cost:**
- Development: Low (reuse existing stack)
- Infrastructure: Same as web (hosting + S3)
- Distribution: Free (no app stores)

---

### React Native (Native Mobile App)

#### What It Provides

**Native Features:**
- Full camera control with `expo-camera`
- Direct photo library access with metadata via `expo-media-library`
- Background upload tasks via `expo-task-manager`
- Full push notification support (iOS + Android)
- Native gestures and animations (smoother UX)
- Better offline support with large storage limits
- System share sheet integration
- Photo widget potential

**Performance:**
- Native rendering (60 FPS animations)
- Better memory management for large image collections
- Faster image processing on device
- Better battery efficiency

**User Experience:**
- App store presence (discovery + trust)
- Native UI patterns (iOS/Android specific)
- Better accessibility
- System-level integration

#### Limitations

**Development Complexity:**
- Separate codebase from web app (60-70% code reuse at best)
- Need to maintain both web and mobile
- Different navigation (React Navigation vs React Router)
- Different styling approach (StyleSheet vs Tailwind)
- Different build and deployment processes

**Distribution:**
- iOS: App Store review process (1-7 days per release)
- Android: Google Play review (faster, but still manual)
- TestFlight for iOS beta testing
- App signing and certificate management

**Platform-Specific Issues:**
- iOS and Android API differences
- Different permission models
- OS version fragmentation
- Need physical devices or emulators for testing

#### Development Requirements

**Technology:**
- Expo (managed workflow) recommended
- React Navigation for routing
- Zustand (works in React Native)
- `expo-image-picker`, `expo-media-library`, `expo-file-system`
- Separate builds for iOS and Android

**Effort Estimate:**
- 8-12 weeks for initial app (with code reuse)
- Ongoing maintenance for both platforms
- Separate release cycles

**Cost:**
- Development: High (separate app)
- Infrastructure: Same as web + push notification service
- Distribution: $99/year (Apple Developer), $25 one-time (Google Play)
- Build services: Expo EAS $29/month (optional)

---

## Feature Comparison for Photo Management Use Case

| Feature | PWA | React Native | Priority |
|---------|-----|--------------|----------|
| **Upload from Camera** | ✅ Good (via file input) | ✅ Excellent (full control) | High |
| **Upload from Library** | ✅ Good (file picker) | ✅ Excellent (with EXIF) | High |
| **Batch Upload** | ✅ Yes | ✅ Yes | High |
| **Offline Browsing** | ✅ Good (with service worker) | ✅ Excellent (more storage) | High |
| **Offline Upload Queue** | ⚠️ Limited (IndexedDB quota) | ✅ Excellent (unlimited) | Medium |
| **Background Sync** | ⚠️ Limited (iOS restrictions) | ✅ Yes (with task manager) | Medium |
| **Push Notifications** | ⚠️ Limited on iOS | ✅ Full support | Low |
| **Image Lazy Loading** | ✅ Yes | ✅ Yes (react-native-fast-image) | High |
| **Tagging Interface** | ✅ Yes | ✅ Yes | High |
| **Search & Filter** | ✅ Yes | ✅ Yes | High |
| **Photo Metadata (EXIF)** | ❌ No client access | ✅ Yes (expo-media-library) | Low |
| **Geolocation Tagging** | ✅ Yes (Geolocation API) | ✅ Yes (Expo Location) | Low |
| **Image Editing** | ⚠️ Limited (canvas-based) | ✅ Better libraries available | V2+ |
| **Face Detection** | ❌ Requires backend | ⚠️ Possible with ML Kit | V2+ |
| **Home Screen Widget** | ❌ No | ✅ Yes (native only) | V2+ |
| **App Store Discovery** | ❌ No | ✅ Yes | Low |
| **Single Codebase** | ✅ Yes | ❌ Separate from web | High |

---

## Recommendation Framework

### Choose PWA if:
✅ Quick time to market is priority (V1)  
✅ Team is already React web developers  
✅ Want to avoid app store complexities  
✅ Budget constraints favor single codebase  
✅ Features needed are mostly CRUD (upload, browse, tag, search)  
✅ Users have modern smartphones with good browsers  
✅ Offline features are "nice to have" not critical  

### Choose React Native if:
✅ Native performance is critical  
✅ Need deep photo library integration with metadata  
✅ Background upload sync is essential  
✅ Push notifications are core feature  
✅ Budget allows for dedicated mobile development  
✅ Team has React Native experience  
✅ App store presence provides value  

---

## Recommended Phased Approach

### Phase 1: PWA (V1 - Months 1-3)

**Goal:** Deliver functional mobile web experience with core features

**Features:**
- Responsive web design (mobile-first)
- PWA manifest and service worker
- Camera and photo library upload
- Offline image browsing (cached thumbnails)
- Basic offline upload queue (IndexedDB)
- Install prompt
- Optimized image loading (lazy + virtualization)

**Benefits:**
- Fast development (extends existing web app)
- Single codebase to maintain
- Immediate deployment (no app store)
- User testing and feedback collection

**Success Metrics:**
- Mobile usage percentage
- PWA install rate
- Upload success rate on mobile
- User feedback on missing native features

### Phase 2: PWA Enhancements (V1.5 - Months 4-5)

**Goal:** Improve mobile experience based on user feedback

**Potential Additions:**
- Push notifications (if users request)
- Better offline support (larger cache quotas)
- Progressive image enhancement
- Mobile-specific UI refinements
- Share API integration
- Camera enhancements (if needed)

**Decision Point:**
Analyze metrics:
- Are users frustrated by PWA limitations?
- Is background upload sync frequently requested?
- Do users want better photo library integration?
- Is performance satisfactory?

### Phase 3: Evaluate React Native (V2 - Month 6+)

**Triggers for Considering React Native:**
- User complaints about PWA limitations (background sync, camera quality)
- Need for photo library metadata access
- Performance issues with large collections on mobile
- Request for app store presence
- Budget available for dedicated mobile development

**If Moving to React Native:**
1. **Shared Package Structure:**
```
packages/
├── shared/           # 60-70% code reuse
│   ├── api/         # API clients (work in both)
│   ├── stores/      # Zustand stores (work in both)
│   ├── utils/       # Validation, formatting, business logic
│   └── types/       # TypeScript definitions
├── web/             # Existing Vite React app
└── mobile/          # New Expo React Native app
```

2. **Incremental Migration:**
   - Keep PWA running (web + mobile web users)
   - Build React Native alongside (for app store users)
   - Share backend API (no changes needed)
   - Gradually migrate power users to native app

3. **Development Timeline:**
   - Weeks 1-2: Project setup, shared package extraction
   - Weeks 3-6: Core screens (browse, upload, detail)
   - Weeks 7-8: Native features (camera, background sync)
   - Weeks 9-10: Testing and refinement
   - Weeks 11-12: App store submission and release

---

## Code Sharing Strategy (if React Native is adopted)

### Shared Code (60-70%)

**API Clients:**
```typescript
// packages/shared/api/images.ts
// Works in both web and React Native
export const fetchImages = async (filters: ImageFilters) => {
  const response = await fetch(`${API_URL}/api/v1/images`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(filters),
  })
  return response.json()
}
```

**Zustand Stores:**
```typescript
// packages/shared/stores/authStore.ts
// Works in both with platform-specific persistence
import { create } from 'zustand'
import AsyncStorage from '@react-native-async-storage/async-storage' // React Native
// import localStorage from 'localStorage' // Web

export const useAuthStore = create((set) => ({
  // Store logic works in both environments
}))
```

**Business Logic:**
```typescript
// packages/shared/utils/validation.ts
import { z } from 'zod'

// Zod schemas work in both environments
export const uploadSchema = z.object({
  tags: z.array(z.string()),
  occasion: z.enum(['birthday', 'wedding', 'vacation', /* ... */]),
})
```

### Platform-Specific Code (30-40%)

**Web (Vite + React):**
- Tailwind CSS styling
- React Router navigation
- `react-dropzone` for file upload
- Service Worker / PWA features

**Mobile (Expo + React Native):**
- StyleSheet or Nativewind styling
- React Navigation
- `expo-image-picker` for camera/library
- Native background tasks

---

## Cost-Benefit Analysis

### PWA Total Cost of Ownership (3 months)

**Development:**
- Initial PWA setup: 40 hours @ $100/hr = $4,000
- Mobile optimizations: 80 hours @ $100/hr = $8,000
- Testing and refinement: 40 hours @ $100/hr = $4,000
- **Total Development: $16,000**

**Ongoing:**
- Maintenance: Same as web (no additional cost)
- Hosting: Same as web (no additional cost)
- **Total Ongoing: $0/month additional**

**Total 1st Year: $16,000**

### React Native Total Cost of Ownership (3 months)

**Development:**
- Project setup and shared packages: 40 hours @ $100/hr = $4,000
- Core app development: 240 hours @ $100/hr = $24,000
- Platform-specific features: 80 hours @ $100/hr = $8,000
- Testing (iOS + Android): 80 hours @ $100/hr = $8,000
- App store setup and submission: 20 hours @ $100/hr = $2,000
- **Total Development: $46,000**

**Ongoing:**
- Maintenance (both platforms): ~40 hours/month @ $100/hr = $4,000/month
- Expo EAS builds: $29/month
- Apple Developer Program: $99/year
- Google Play Console: $25 one-time
- Push notification service: ~$20/month
- **Total Ongoing: ~$4,150/month**

**Total 1st Year: $96,000**

### Savings with PWA First Approach

- **Immediate savings: $30,000** (development)
- **Ongoing savings: $4,150/month** ($50,000/year)
- **Faster time to market:** 2-3 months sooner
- **Risk reduction:** Validate product-market fit before native investment

---

## User Experience Comparison

### PWA Experience

**Installation:**
1. User visits website on mobile
2. Browser shows "Add to Home Screen" prompt
3. User adds to home screen (optional)
4. Icon appears on launcher

**Upload Flow:**
1. Open app (web or installed)
2. Tap "Upload" button
3. Choose "Take Photo" or "Choose from Library"
4. Native camera/picker opens
5. Select photos (multi-select supported)
6. Add tags and metadata
7. Upload (or queue if offline)

**Browsing:**
- Grid view with lazy-loaded thumbnails
- Smooth scrolling with virtualization
- Tap to view detail
- Swipe gestures (via web APIs)
- Search and filter

**Offline:**
- Browse cached thumbnails
- Queue uploads for sync
- "You're offline" indicator
- Auto-sync when connection restored

### React Native Experience

**Installation:**
1. Download from App Store / Google Play
2. Install (standard app installation)
3. Open from launcher

**Upload Flow:**
1. Open app
2. Tap "Upload" button
3. Native image picker with full features
   - See all albums
   - Multi-select with preview
   - Access photo metadata (date, location)
4. Photos load instantly (no upload to view)
5. Add tags and metadata
6. Upload in background (even when app closed)

**Browsing:**
- Native list/grid components
- Smooth 60 FPS scrolling
- Native gestures (swipe, pinch-to-zoom)
- System-level integration
- Faster perceived performance

**Offline:**
- Full offline capability with large storage
- Background sync even when app closed
- System notifications for sync progress
- Better battery management

---

## Technical Feasibility Assessment

### PWA Readiness for V1 ✅

**Current State:**
- Existing React web app provides foundation
- Responsive design framework (Tailwind) ready
- Image optimization already planned
- API already designed for mobile consumption

**Remaining Work:**
- Add PWA manifest and service worker (2-3 days)
- Implement offline storage with IndexedDB (3-5 days)
- Mobile-optimize upload flow (5-7 days)
- Add camera/library access (2-3 days)
- Test on iOS and Android devices (5-7 days)
- Refine mobile UI/UX (5-10 days)

**Total Estimated Time: 3-5 weeks**

### React Native Readiness (Future) ⏳

**Prerequisites:**
- Extract shared business logic to packages
- Define platform-specific UI patterns
- Set up build infrastructure (Expo EAS or custom)
- Acquire iOS Developer Account ($99/year)
- Acquire Google Play Account ($25 one-time)
- Establish React Native development environment

**Estimated Time: 8-12 weeks** (after prerequisites)

---

## Decision Matrix

| Criteria | Weight | PWA Score | React Native Score | PWA Weighted | RN Weighted |
|----------|--------|-----------|-------------------|--------------|-------------|
| Time to Market | 25% | 9/10 | 5/10 | 2.25 | 1.25 |
| Development Cost | 20% | 9/10 | 4/10 | 1.80 | 0.80 |
| Maintenance Burden | 15% | 9/10 | 5/10 | 1.35 | 0.75 |
| Feature Completeness | 15% | 7/10 | 9/10 | 1.05 | 1.35 |
| User Experience | 10% | 7/10 | 9/10 | 0.70 | 0.90 |
| Offline Capability | 10% | 7/10 | 9/10 | 0.70 | 0.90 |
| Scalability | 5% | 8/10 | 8/10 | 0.40 | 0.40 |
| **Total** | **100%** | | | **8.25** | **6.35** |

**Conclusion:** PWA scores significantly higher for V1 due to time-to-market, cost, and simplicity advantages.

---

## Migration Path (if choosing React Native later)

### Step 1: Prepare Shared Code (1-2 weeks)
- Extract API clients to `packages/shared/api`
- Move Zustand stores to `packages/shared/stores`
- Move validation schemas to `packages/shared/utils`
- Move TypeScript types to `packages/shared/types`

### Step 2: Set Up React Native Project (1 week)
- Initialize Expo managed workflow
- Configure TypeScript
- Set up navigation structure
- Configure build settings

### Step 3: Build Core Screens (3-4 weeks)
- Browse/Grid view with thumbnails
- Image detail view
- Upload flow with camera integration
- Search and filter
- User settings

### Step 4: Native Features (2-3 weeks)
- Background upload with `expo-task-manager`
- Push notifications with `expo-notifications`
- Photo library integration with metadata
- Share extension

### Step 5: Testing & Release (2-3 weeks)
- Test on physical iOS devices
- Test on physical Android devices
- Beta testing with TestFlight / Firebase App Distribution
- App Store submission
- Google Play submission

**Total Timeline: 10-14 weeks**

---

## Recommendation Summary

### For V1 (Next 3-6 months): Build PWA ✅

**Rationale:**
1. **Faster Time to Market:** 3-5 weeks vs 10-14 weeks
2. **Lower Cost:** $16K vs $96K first year
3. **Single Codebase:** Easier maintenance, faster iteration
4. **Sufficient Features:** 90% of needed functionality available
5. **Risk Mitigation:** Validate features before native investment
6. **User Feedback:** Learn what features are actually critical

**What to Build:**
- Responsive, mobile-first web app
- PWA with install prompt and offline support
- Camera and photo library upload
- Optimized image loading and caching
- Mobile-optimized upload workflow
- IndexedDB offline queue

### For V2 (6+ months): Evaluate React Native

**Decision Criteria:**
- Analyze V1 mobile usage and satisfaction metrics
- Identify feature gaps based on user feedback
- Assess budget for native development
- Evaluate if background sync is critical
- Consider app store presence value

**If Moving to React Native:**
- Extract shared code first
- Build React Native alongside PWA
- Keep both versions running initially
- Gradually migrate power users to native

---

## Next Steps

1. **Proceed with PWA for V1**
   - Add PWA features to existing Vite app
   - Test on iOS Safari and Android Chrome
   - Deploy with install prompt

2. **Gather User Feedback**
   - Track mobile usage analytics
   - Survey users about experience
   - Identify pain points and feature requests

3. **Review After 3-6 Months**
   - Assess PWA performance and satisfaction
   - Determine if React Native is needed
   - Plan migration if justified by data

**Conclusion:** PWA provides the optimal balance of functionality, cost, and time-to-market for V1. React Native remains a viable option for V2 if user needs justify the investment.

---

## Testing Strategy for Mobile

### PWA Testing Approach

#### 1. Mobile Browser Testing

**Test Matrix:**
| Device | Browser | Priority |
|--------|---------|----------|
| iPhone (iOS 16+) | Safari | High |
| iPhone (iOS 15) | Safari | Medium |
| Android (latest) | Chrome | High |
| Android (latest) | Firefox | Low |
| iPad | Safari | Medium |

**Key Test Scenarios:**
- PWA installation flow
- Camera/photo library access
- Offline functionality
- Touch gestures (swipe, pinch, long-press)
- Service worker caching
- IndexedDB storage limits
- Push notifications (where supported)

#### 2. Playwright Mobile Testing

**Mobile Configuration:**
```typescript
// playwright.config.ts
export default defineConfig({
  projects: [
    {
      name: 'Mobile Chrome',
      use: {
        ...devices['Pixel 5'],
        viewport: { width: 393, height: 851 },
        hasTouch: true,
        isMobile: true,
      },
    },
    {
      name: 'Mobile Safari',
      use: {
        ...devices['iPhone 13'],
        viewport: { width: 390, height: 844 },
        hasTouch: true,
        isMobile: true,
      },
    },
  ],
})
```

**Mobile-Specific E2E Tests:**
```typescript
// e2e/mobile-upload.spec.ts
import { test, expect } from '@playwright/test'

test.describe('Mobile Upload Flow', () => {
  test.use({ ...devices['iPhone 13'] })

  test('uploads photo from camera', async ({ page }) => {
    await page.goto('/upload')
    
    // Intercept file input (camera)
    const fileInput = page.locator('input[type="file"][capture="environment"]')
    await fileInput.setInputFiles('fixtures/test-photo.jpg')
    
    await expect(page.getByText('1 photo ready')).toBeVisible()
    
    // Test touch interactions
    await page.tap('[data-testid="tag-input"]')
    await page.fill('[data-testid="tag-input"]', 'vacation')
    await page.tap('[data-testid="add-tag-button"]')
    
    await page.tap('[data-testid="upload-button"]')
    await expect(page.getByText('Upload complete')).toBeVisible({ timeout: 10000 })
  })

  test('works offline', async ({ page, context }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    // Go offline
    await context.setOffline(true)
    
    // Should still show cached images
    await expect(page.locator('[data-testid="image-card"]').first()).toBeVisible()
    
    // Upload should queue
    await page.goto('/upload')
    const fileInput = page.locator('input[type="file"]')
    await fileInput.setInputFiles('fixtures/test-photo.jpg')
    
    await page.tap('[data-testid="upload-button"]')
    await expect(page.getByText(/will upload when online/i)).toBeVisible()
  })

  test('responsive layout on mobile', async ({ page }) => {
    await page.goto('/')
    
    // Check mobile menu
    await expect(page.locator('[data-testid="mobile-menu"]')).toBeVisible()
    await expect(page.locator('[data-testid="desktop-menu"]')).not.toBeVisible()
    
    // Check image grid adapts
    const grid = page.locator('[data-testid="image-grid"]')
    const boundingBox = await grid.boundingBox()
    expect(boundingBox?.width).toBeLessThan(400)
  })
})
```

#### 3. Real Device Testing

**Using BrowserStack or Sauce Labs:**
```typescript
// playwright.config.ts (BrowserStack integration)
import { defineConfig } from '@playwright/test'

export default defineConfig({
  use: {
    connectOptions: {
      wsEndpoint: `wss://cdp.browserstack.com/playwright?caps=${encodeURIComponent(JSON.stringify({
        'browser': 'chrome',
        'os': 'android',
        'os_version': '13.0',
        'browserstack.username': process.env.BROWSERSTACK_USERNAME,
        'browserstack.accessKey': process.env.BROWSERSTACK_ACCESS_KEY,
      }))}`,
    },
  },
})
```

**Manual Testing Checklist:**
```markdown
# Mobile PWA Testing Checklist

## Installation
- [ ] "Add to Home Screen" prompt appears
- [ ] App icon displays correctly
- [ ] Splash screen shows on launch
- [ ] App opens in fullscreen mode

## Camera/Upload
- [ ] Camera permission requested
- [ ] Camera opens and captures photo
- [ ] Photo library accessible
- [ ] Multi-select works
- [ ] Preview images load

## Offline
- [ ] Service worker registers
- [ ] Cached images display offline
- [ ] Upload queue works offline
- [ ] Syncs when back online
- [ ] "Offline" indicator shows

## Performance
- [ ] Images lazy load smoothly
- [ ] Scrolling is smooth (60fps)
- [ ] Touch gestures responsive
- [ ] No layout shift on load

## Accessibility
- [ ] Touch targets ≥44x44px
- [ ] Zoom works properly
- [ ] Screen reader navigation
- [ ] Color contrast sufficient
```

### React Native Testing (if adopted in V2)

#### 1. Jest + React Native Testing Library

**Configuration:**
```javascript
// jest.config.js
module.exports = {
  preset: 'react-native',
  setupFilesAfterEnv: ['@testing-library/jest-native/extend-expect'],
  transformIgnorePatterns: [
    'node_modules/(?!(react-native|@react-native|expo|@expo)/)',
  ],
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.stories.tsx',
    '!src/**/*.test.tsx',
  ],
}
```

**Example Test:**
```typescript
// components/MobileImageGrid.test.tsx
import { render, screen } from '@testing-library/react-native'
import { MobileImageGrid } from './MobileImageGrid'

describe('MobileImageGrid', () => {
  it('renders image thumbnails', () => {
    const images = [
      { id: '1', thumbnailUrl: 'https://example.com/1.jpg' },
      { id: '2', thumbnailUrl: 'https://example.com/2.jpg' },
    ]
    
    render(<MobileImageGrid images={images} />)
    
    expect(screen.getByTestId('image-grid')).toBeTruthy()
    expect(screen.getAllByTestId('image-thumbnail')).toHaveLength(2)
  })
})
```

#### 2. Detox (E2E for React Native)

**Configuration:**
```json
// .detoxrc.js
module.exports = {
  testRunner: {
    args: {
      config: 'e2e/jest.config.js',
    },
    jest: {
      setupTimeout: 120000,
    },
  },
  apps: {
    'ios.release': {
      type: 'ios.app',
      binaryPath: 'ios/build/Build/Products/Release-iphonesimulator/PhotoManager.app',
      build: 'xcodebuild -workspace ios/PhotoManager.xcworkspace -scheme PhotoManager -configuration Release -sdk iphonesimulator -derivedDataPath ios/build',
    },
    'android.release': {
      type: 'android.apk',
      binaryPath: 'android/app/build/outputs/apk/release/app-release.apk',
      build: 'cd android && ./gradlew assembleRelease assembleAndroidTest -DtestBuildType=release',
    },
  },
  devices: {
    simulator: {
      type: 'ios.simulator',
      device: { type: 'iPhone 14' },
    },
    emulator: {
      type: 'android.emulator',
      device: { avdName: 'Pixel_5_API_31' },
    },
  },
}
```

**Example Detox Test:**
```typescript
// e2e/upload.e2e.ts
import { device, element, by, expect as detoxExpect } from 'detox'

describe('Upload Flow', () => {
  beforeAll(async () => {
    await device.launchApp()
  })

  it('should upload photo from camera', async () => {
    await element(by.id('upload-tab')).tap()
    await element(by.id('camera-button')).tap()
    
    // Grant camera permission if needed
    await device.takeScreenshot('camera-permission')
    
    // Simulate photo capture (requires mock)
    await element(by.id('capture-button')).tap()
    
    // Add tags
    await element(by.id('tag-input')).typeText('vacation')
    await element(by.id('add-tag-button')).tap()
    
    // Upload
    await element(by.id('upload-button')).tap()
    
    await detoxExpect(element(by.text('Upload complete'))).toBeVisible()
  })
})
```

### Mobile Testing Best Practices

#### 1. Test on Real Devices

**Why:**
- Emulators don't accurately simulate:
  - Touch responsiveness
  - Camera quality
  - Battery usage
  - Network conditions
  - Performance constraints

**Minimum Device Coverage:**
- 1x iPhone (latest iOS)
- 1x iPhone (iOS - 1 version)
- 1x Android (flagship, latest)
- 1x Android (mid-range, 2 years old)

#### 2. Network Condition Testing

```typescript
// Playwright network throttling
test('handles slow network', async ({ page, context }) => {
  // Simulate 3G
  await context.route('**/*', route => {
    setTimeout(() => route.continue(), 500) // 500ms delay
  })
  
  await page.goto('/')
  
  // Should show loading states
  await expect(page.getByTestId('skeleton-loader')).toBeVisible()
})

test('handles offline gracefully', async ({ page, context }) => {
  await page.goto('/')
  await page.waitForLoadState('networkidle')
  
  await context.setOffline(true)
  
  // Try to upload
  await page.goto('/upload')
  await expect(page.getByText(/offline/i)).toBeVisible()
})
```

#### 3. Touch Gesture Testing

```typescript
// Test swipe gestures
test('swipes to delete image', async ({ page }) => {
  await page.goto('/gallery')
  
  const imageCard = page.locator('[data-testid="image-card"]').first()
  const box = await imageCard.boundingBox()
  
  // Swipe left
  await page.touchscreen.tap(box.x + box.width / 2, box.y + box.height / 2)
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2)
  await page.mouse.down()
  await page.mouse.move(box.x - 100, box.y + box.height / 2)
  await page.mouse.up()
  
  await expect(page.getByRole('button', { name: 'Delete' })).toBeVisible()
})
```

### Maintainability for Mobile

#### 1. Feature Flags

```typescript
// utils/features.ts
export const features = {
  pwaInstallPrompt: true,
  offlineMode: true,
  cameraUpload: true,
  pushNotifications: false, // Not ready yet
}

// Usage in component
if (features.pwaInstallPrompt) {
  return <InstallPrompt />
}
```

#### 2. Platform Detection

```typescript
// utils/platform.ts
export const isMobile = () => {
  return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
    navigator.userAgent
  )
}

export const isIOS = () => {
  return /iPhone|iPad|iPod/i.test(navigator.userAgent)
}

export const isAndroid = () => {
  return /Android/i.test(navigator.userAgent)
}

export const isPWA = () => {
  return window.matchMedia('(display-mode: standalone)').matches
}

// Usage
if (isIOS() && !isPWA()) {
  return <IOSInstallInstructions />
}
```

#### 3. Responsive Testing Utilities

```typescript
// test/utils/responsive.tsx
import { render } from '@testing-library/react'

export const renderMobile = (ui: React.ReactElement) => {
  // Mock mobile viewport
  Object.defineProperty(window, 'innerWidth', { value: 375 })
  Object.defineProperty(window, 'innerHeight', { value: 667 })
  
  window.matchMedia = vi.fn().mockImplementation((query) => ({
    matches: query.includes('max-width: 768px'),
    media: query,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  }))
  
  return render(ui)
}

// Usage in tests
it('renders mobile layout', () => {
  renderMobile(<ImageGrid />)
  expect(screen.getByTestId('mobile-grid')).toBeInTheDocument()
})
```

#### 4. PWA Testing Utilities

```typescript
// test/utils/pwa.ts
export const mockServiceWorker = () => {
  global.navigator.serviceWorker = {
    register: vi.fn().mockResolvedValue({
      active: { postMessage: vi.fn() },
    }),
    ready: Promise.resolve({
      active: { postMessage: vi.fn() },
    }),
  } as any
}

export const mockBeforeInstallPrompt = () => {
  const prompt = vi.fn()
  const event = {
    preventDefault: vi.fn(),
    prompt,
    userChoice: Promise.resolve({ outcome: 'accepted' }),
  }
  
  window.dispatchEvent(new Event('beforeinstallprompt'))
  return event
}
```

### Mobile CI/CD Pipeline

```yaml
# .github/workflows/mobile-test.yml
name: Mobile Tests

on: [push, pull_request]

jobs:
  pwa-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      
      - name: Install dependencies
        run: npm ci
      
      - name: Run unit tests
        run: npm run test:coverage
      
      - name: Install Playwright browsers
        run: npx playwright install --with-deps
      
      - name: Run Playwright mobile tests
        run: npm run test:e2e -- --project="Mobile Chrome" --project="Mobile Safari"
      
      - name: Upload test results
        if: failure()
        uses: actions/upload-artifact@v3
        with:
          name: playwright-report
          path: playwright-report/
  
  browserstack-test:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      
      - name: Test on real iOS devices
        env:
          BROWSERSTACK_USERNAME: ${{ secrets.BROWSERSTACK_USERNAME }}
          BROWSERSTACK_ACCESS_KEY: ${{ secrets.BROWSERSTACK_ACCESS_KEY }}
        run: npm run test:browserstack
```

---

## Summary: Testing Recommendations

### For PWA (V1):
1. **Vitest** for unit/integration tests (fast, Vite-native)
2. **React Testing Library** for component tests (user-centric)
3. **Playwright** for E2E with mobile device emulation
4. **Storybook** for component development and visual testing
5. **Manual testing** on 2-3 real devices (iPhone + Android)
6. **BrowserStack** for broader device coverage (optional, V1.5+)

### For React Native (V2 if needed):
1. **Jest** + **React Native Testing Library** for components
2. **Detox** for E2E on iOS/Android simulators
3. **Physical device testing** before releases
4. **TestFlight** (iOS) and **Firebase App Distribution** (Android) for beta testing

### Team Efficiency:
- ✅ **Single test framework family** (Vitest/Jest share syntax)
- ✅ **Minimal configuration** (Vite integration just works)
- ✅ **Fast feedback** (instant test reruns)
- ✅ **Shared knowledge** (same patterns for web and potential native)

**Total testing setup time:** 2-3 days for PWA, additional 1-2 weeks for comprehensive E2E suite.
