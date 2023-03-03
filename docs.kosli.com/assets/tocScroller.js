document.addEventListener('DOMContentLoaded', function(e) {
    Scroller.init();
});

class Scroller {
    static init() {
      if(document.querySelector('.book-toc-content').hasChildNodes()) {
        this.tocLinks = document.querySelectorAll('.book-toc-content a');
        this.tocLinks[0].classList.add('active');
        this.headers = Array.from(this.tocLinks).map(link => {
            return document.querySelector(`#${link.href.split('#')[1]}`);
        })
        this.ticking = false;
        window.addEventListener('scroll', (e) => {
          this.onScroll()
        })
      }
    }
  
    static onScroll() {
      if(!this.ticking) {
        requestAnimationFrame(this.update.bind(this));
        this.ticking = true;
      }
    }
  
    static update() {
      this.activeHeader ||= this.headers[0];
      let activeIndex = this.headers.findIndex((header) => {
        return header.getBoundingClientRect().top > 180;
      });
      if(activeIndex == -1) {
        activeIndex = this.headers.length - 1;
      } else if(activeIndex > 0) {
        activeIndex--;
      }
      let active = this.headers[activeIndex];
      if(active !== this.activeHeader) {
        this.activeHeader = active;
        this.tocLinks.forEach(link => link.classList.remove('active'));
        this.tocLinks[activeIndex].classList.add('active');
      }
      this.ticking = false;
    }
}