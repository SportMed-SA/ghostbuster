import { Component, inject, signal } from '@angular/core';
import { TranslatePipe, TranslateService } from '@ngx-translate/core';

@Component({
  selector: 'app-root',
  imports: [TranslatePipe],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App {
  protected readonly currentLang = signal('en');

  private readonly translate = inject(TranslateService);

  constructor() {
    this.translate.addLangs(['en', 'de']);
    this.translate.setFallbackLang('en');
    this.translate.use('en');
  }

  protected switchLanguage(lang: string): void {
    this.currentLang.set(lang);
    this.translate.use(lang);
  }
}
