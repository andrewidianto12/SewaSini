document.addEventListener('DOMContentLoaded', function() {
    // Room type filtering
    const filterButtons = document.querySelectorAll('.filter-btn');
    const roomItems = document.querySelectorAll('.room-item');

    filterButtons.forEach(function(btn) {
        btn.addEventListener('click', function() {
            filterButtons.forEach(function(b) {
                b.classList.remove('active', 'btn-primary');
                b.classList.add('btn-outline-primary');
            });
            this.classList.add('active', 'btn-primary');
            this.classList.remove('btn-outline-primary');

            var type = this.getAttribute('data-type');
            roomItems.forEach(function(item) {
                if (type === 'all' || item.getAttribute('data-type') === type) {
                    item.style.display = '';
                } else {
                    item.style.display = 'none';
                }
            });
        });
    });

    // Date constraints
    var startDate = document.getElementById('startDate');
    var endDate = document.getElementById('endDate');

    if (startDate && endDate) {
        var today = new Date().toISOString().split('T')[0];
        startDate.setAttribute('min', today);
        endDate.setAttribute('min', today);

        startDate.addEventListener('change', function() {
            endDate.setAttribute('min', this.value);
            if (endDate.value && endDate.value < this.value) {
                endDate.value = this.value;
            }
        });
    }

    // Form validation
    var forms = document.querySelectorAll('#bookingForm');
    forms.forEach(function(form) {
        form.addEventListener('submit', function(e) {
            if (startDate && endDate) {
                if (endDate.value < startDate.value) {
                    e.preventDefault();
                    alert('Tanggal selesai harus setelah atau sama dengan tanggal mulai!');
                }
            }
        });
    });
});
