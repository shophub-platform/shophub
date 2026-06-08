import { Component, OnDestroy, OnInit, signal } from '@angular/core';

interface Slide {
  icon: string;
  title: string;
  text: string;
  gradient: string;
}

/** Auto-rotirajući hero slajdšou (cross-fade) — dekorativni banner na dashboard-u. */
@Component({
  selector: 'app-hero-carousel',
  standalone: true,
  template: `
    <div class="hero">
      @for (slide of slides; track slide.title; let i = $index) {
        <div
          class="slide"
          [style.background]="slide.gradient"
          [class.active]="i === current()"
          [attr.aria-hidden]="i === current() ? null : true"
        >
          <span class="material-icons hero-icon">{{ slide.icon }}</span>
          <div class="hero-text">
            <h3>{{ slide.title }}</h3>
            <p>{{ slide.text }}</p>
          </div>
        </div>
      }

      <div class="dots">
        @for (slide of slides; track slide.title; let i = $index) {
          <button
            class="dot"
            [class.on]="i === current()"
            (click)="go(i)"
            [attr.aria-label]="'Slide ' + (i + 1)"
          ></button>
        }
      </div>
    </div>
  `,
  styles: [
    `
      .hero {
        position: relative;
        height: 150px;
        border-radius: 18px;
        overflow: hidden;
        margin-bottom: 24px;
        box-shadow: 0 10px 30px rgba(16, 24, 40, 0.18);
      }
      .slide {
        position: absolute;
        inset: 0;
        display: flex;
        align-items: center;
        gap: 20px;
        padding: 0 32px;
        color: #fff;
        opacity: 0;
        transform: scale(1.04);
        transition: opacity 0.8s ease, transform 0.8s ease;
        pointer-events: none;
      }
      .slide.active {
        opacity: 1;
        transform: scale(1);
        pointer-events: auto;
      }
      .hero-icon {
        font-size: 52px;
        opacity: 0.95;
        filter: drop-shadow(0 4px 10px rgba(0, 0, 0, 0.25));
      }
      .hero-text h3 {
        margin: 0;
        font-size: 22px;
        font-weight: 700;
        letter-spacing: 0.2px;
      }
      .hero-text p {
        margin: 4px 0 0;
        opacity: 0.92;
        font-size: 14px;
      }
      .dots {
        position: absolute;
        bottom: 12px;
        right: 16px;
        display: flex;
        gap: 6px;
      }
      .dot {
        width: 8px;
        height: 8px;
        border-radius: 999px;
        border: none;
        background: rgba(255, 255, 255, 0.5);
        cursor: pointer;
        padding: 0;
        transition: width 0.3s ease, background 0.3s ease;
      }
      .dot.on {
        width: 22px;
        background: #fff;
      }
    `,
  ],
})
export class HeroCarouselComponent implements OnInit, OnDestroy {
  readonly current = signal(0);
  private timer?: ReturnType<typeof setInterval>;

  readonly slides: Slide[] = [
    {
      icon: 'rocket_launch',
      title: 'Deploy in seconds',
      text: 'Spin up shops backed by Kubernetes.',
      gradient: 'linear-gradient(135deg, #4f46e5, #7c3aed)',
    },
    {
      icon: 'verified_user',
      title: 'Secure by default',
      text: 'JWT auth and isolated per-shop resources.',
      gradient: 'linear-gradient(135deg, #0ea5e9, #2563eb)',
    },
    {
      icon: 'insights',
      title: 'Full visibility',
      text: 'Track status from Provisioning to Ready.',
      gradient: 'linear-gradient(135deg, #059669, #0d9488)',
    },
  ];

  ngOnInit(): void {
    this.timer = setInterval(() => this.next(), 4500);
  }

  ngOnDestroy(): void {
    if (this.timer) {
      clearInterval(this.timer);
    }
  }

  next(): void {
    this.current.update((i) => (i + 1) % this.slides.length);
  }

  go(i: number): void {
    this.current.set(i);
  }
}
